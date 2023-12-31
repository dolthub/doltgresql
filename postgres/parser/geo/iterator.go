// Copyright 2023 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geo

import (
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

// GeomTIterator decomposes geom.T  objects into individual components
// (i.e. either a POINT, LINESTRING or POLYGON) and presents them as an
// iterator. It automatically decomposes MULTI-* and GEOMETRYCOLLECTION
// objects. This prevents an allocation compared to decomposing the objects
// into a geom.T array.
type GeomTIterator struct {
	g             geom.T
	emptyBehavior EmptyBehavior
	// idx is the index into the MULTI-* or GEOMETRYCOLLECTION item.
	idx int
	// subIt is the iterator inside a GeometryCollection.
	// Note GeometryCollections can be nested.
	subIt *GeomTIterator
}

// NewGeomTIterator returns a new GeomTIterator.
func NewGeomTIterator(g geom.T, emptyBehavior EmptyBehavior) GeomTIterator {
	return GeomTIterator{g: g, emptyBehavior: emptyBehavior}
}

// Next returns the next geom.T object, a bool as to whether there is an
// entry and an error if any.
func (it *GeomTIterator) Next() (geom.T, bool, error) {
	next, hasNext, err := it.next()
	if err != nil || !hasNext {
		return nil, hasNext, err
	}
	for {
		if !next.Empty() {
			return next, hasNext, nil
		}
		switch it.emptyBehavior {
		case EmptyBehaviorOmit:
			next, hasNext, err = it.next()
			if err != nil || !hasNext {
				return nil, hasNext, err
			}
		case EmptyBehaviorError:
			return nil, false, NewEmptyGeometryError()
		default:
			return nil, false, errors.Newf("programmer error: unknown behavior: %T", it.emptyBehavior)
		}
	}
}

// next() is the internal method for Next.
// It handles fetching the next item in the iterator, recursing down structures
// if necessary. It does not check for emptiness.
func (it *GeomTIterator) next() (geom.T, bool, error) {
	switch t := it.g.(type) {
	case *geom.Point, *geom.LineString, *geom.Polygon:
		if it.idx == 1 {
			return nil, false, nil
		}
		it.idx++
		return t, true, nil
	case *geom.MultiPoint:
		if it.idx == t.NumPoints() {
			return nil, false, nil
		}
		p := t.Point(it.idx)
		it.idx++
		return p, true, nil
	case *geom.MultiLineString:
		if it.idx == t.NumLineStrings() {
			return nil, false, nil
		}
		p := t.LineString(it.idx)
		it.idx++
		return p, true, nil
	case *geom.MultiPolygon:
		if it.idx == t.NumPolygons() {
			return nil, false, nil
		}
		p := t.Polygon(it.idx)
		it.idx++
		return p, true, nil
	case *geom.GeometryCollection:
		for {
			if it.idx == t.NumGeoms() {
				return nil, false, nil
			}
			if it.subIt == nil {
				it.subIt = &GeomTIterator{g: t.Geom(it.idx), emptyBehavior: it.emptyBehavior}
			}
			ret, next, err := it.subIt.next()
			if err != nil {
				return nil, false, err
			}
			if next {
				return ret, next, nil
			}
			// Reset and move to the next item in the collection.
			it.idx++
			it.subIt = nil
		}
	default:
		return nil, false, errors.Newf("unknown type: %T", t)
	}
}
