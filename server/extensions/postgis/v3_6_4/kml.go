// Copyright 2026 Dolthub, Inc.
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

package v3_6_4

import (
	"encoding/xml"
	"io"
	"math"
	"slices"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// kmlNamespace is the namespace of KML elements.
const kmlNamespace = "http://www.opengis.net/kml/2.2"

// errInvalidKML is the error for any KML that cannot be read as a geometry.
var errInvalidKML = pgerror.New(pgcode.Internal, "invalid KML representation")

// xmlNode is an element of an XML document, with its local name, its attributes by local name, its child elements, and
// its text. An element counts as being in the expected namespaces when it has no prefix, when its prefix is
// undeclared, or when its prefix is declared for one of those namespaces.
type xmlNode struct {
	name     string
	inSpace  bool
	attrs    map[string]string
	children []*xmlNode
	text     strings.Builder
	// namespace maps the prefixes declared on the element and its ancestors to their namespaces.
	namespace map[string]string
}

// parseXMLTree reads an XML document that holds a single root element, returning false when the document is
// malformed.
func parseXMLTree(text string, namespaces ...string) (*xmlNode, bool) {
	decoder := xml.NewDecoder(strings.NewReader(text))
	var root *xmlNode
	var stack []*xmlNode
	for {
		token, err := decoder.RawToken()
		if err == io.EOF {
			return root, root != nil && len(stack) == 0
		}
		if err != nil {
			return nil, false
		}
		switch token := token.(type) {
		case xml.StartElement:
			if root != nil && len(stack) == 0 {
				return nil, false
			}
			node := &xmlNode{name: token.Name.Local, attrs: map[string]string{}, namespace: map[string]string{}}
			if len(stack) > 0 {
				for prefix, uri := range stack[len(stack)-1].namespace {
					node.namespace[prefix] = uri
				}
			}
			for _, attr := range token.Attr {
				if attr.Name.Space == "xmlns" {
					node.namespace[attr.Name.Local] = attr.Value
				} else if attr.Name.Space != "" || attr.Name.Local != "xmlns" {
					node.attrs[attr.Name.Local] = attr.Value
				}
			}
			uri, declared := node.namespace[token.Name.Space]
			node.inSpace = token.Name.Space == "" || !declared || slices.Contains(namespaces, uri)
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, node)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) == 0 || stack[len(stack)-1].name != token.Name.Local {
				return nil, false
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].text.Write(token)
			} else if strings.TrimSpace(string(token)) != "" {
				return nil, false
			}
		}
	}
}

// childrenNamed returns the child elements in the expected namespace with the given name.
func (n *xmlNode) childrenNamed(name string) []*xmlNode {
	var found []*xmlNode
	for _, child := range n.children {
		if child.inSpace && child.name == name {
			found = append(found, child)
		}
	}
	return found
}

// stGeomFromKML implements ST_GeomFromKML, which reads a KML geometry in longitude and latitude.
func stGeomFromKML(ctx *sql.Context, args ...any) (any, error) {
	root, ok := parseXMLTree(args[0].(string), kmlNamespace)
	if !ok {
		return nil, errInvalidKML
	}
	s, err := readKML(ctx, root)
	if err != nil {
		return nil, err
	}
	return Geometry{SRID: sridDefault, Shape: s}, nil
}

// readKML reads a KML geometry element. A MultiGeometry is read as a collection of its geometry elements, which is
// then homogenized, and it keeps Z coordinates only when all of its members have them, as a polygon does with its
// rings.
func readKML(ctx *sql.Context, n *xmlNode) (shape, error) {
	if !n.inSpace {
		return shape{}, errInvalidKML
	}
	switch n.name {
	case "Point", "LineString":
		seq, ctype, err := kmlCoordinates(n)
		if err != nil {
			return shape{}, err
		}
		count := len(seq) / ctype.Dimension()
		if (n.name == "Point" && count != 1) || (n.name == "LineString" && count < 2) {
			return shape{}, errInvalidKML
		}
		if n.name == "Point" {
			return shape{typeNum: pointType, ctype: ctype, coords: seq}, nil
		}
		return shape{typeNum: lineType, ctype: ctype, coords: seq}, nil
	case "Polygon":
		outer := n.childrenNamed("outerBoundaryIs")
		if len(outer) != 1 {
			return shape{}, errInvalidKML
		}
		var rings []shape
		for _, boundary := range append(outer, n.childrenNamed("innerBoundaryIs")...) {
			linearRings := boundary.childrenNamed("LinearRing")
			if len(linearRings) == 0 {
				return shape{}, errInvalidKML
			}
			seq, ctype, err := kmlCoordinates(linearRings[0])
			if err != nil {
				return shape{}, err
			}
			dims := ctype.Dimension()
			if len(seq)/dims < 4 {
				return shape{}, errInvalidKML
			}
			if !equalOrdinates(seq[:dims], seq[len(seq)-dims:]) {
				sendNotice(ctx, "NOTICE", "forced closure on an un-closed KML polygon")
				seq = append(seq, seq[:dims]...)
			}
			rings = append(rings, shape{typeNum: lineType, ctype: ctype, coords: seq})
		}
		ctype := commonCoordinatesType(rings)
		polygon := shape{typeNum: polygonType, ctype: ctype}
		for _, ring := range rings {
			polygon.rings = append(polygon.rings, forceDims(ring, ctype, 0, 0).coords)
		}
		return polygon, nil
	case "MultiGeometry":
		var members []shape
		for _, child := range n.children {
			switch child.name {
			case "Point", "LineString", "Polygon", "MultiGeometry":
				if !child.inSpace {
					continue
				}
				member, err := readKML(ctx, child)
				if err != nil {
					return shape{}, err
				}
				members = append(members, member)
			}
		}
		ctype := commonCoordinatesType(members)
		collection := shape{typeNum: collectionType, ctype: ctype}
		for _, member := range members {
			collection.members = append(collection.members, forceDims(member, ctype, 0, 0))
		}
		if collection.isEmpty() {
			return shape{typeNum: collectionType}, nil
		}
		return homogenize(collection), nil
	default:
		return shape{}, errInvalidKML
	}
}

// commonCoordinatesType returns XYZ when every shape has Z coordinates, and XY otherwise.
func commonCoordinatesType(shapes []shape) coordinatesType {
	for _, s := range shapes {
		if !s.hasZ() {
			return dimXY
		}
	}
	if len(shapes) == 0 {
		return dimXY
	}
	return dimXYZ
}

// kmlCoordinates reads the points of an element's first coordinates element. The numbers of a point are separated by
// commas, which may have whitespace around them, while points are separated by whitespace alone. Every point has two
// or three numbers, and all points have the same number.
func kmlCoordinates(n *xmlNode) ([]float64, coordinatesType, error) {
	elements := n.childrenNamed("coordinates")
	if len(elements) == 0 {
		return nil, dimXY, errInvalidKML
	}
	text := elements[0].text.String()
	var points [][]float64
	afterComma := false
	for i := 0; i < len(text); {
		if isCSpace(text[i]) {
			i++
		} else if text[i] == ',' {
			afterComma = true
			i++
		} else {
			match := scanFloat.FindString(text[i:])
			val, err := parseStrtod(match)
			if match == "" || err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
				return nil, dimXY, errInvalidKML
			}
			if len(points) == 0 || !afterComma {
				points = append(points, nil)
			}
			points[len(points)-1] = append(points[len(points)-1], val)
			afterComma = false
			i += len(match)
		}
	}
	if afterComma || len(points) == 0 {
		return nil, dimXY, errInvalidKML
	}
	var seq []float64
	for _, point := range points {
		if len(point) < 2 || len(point) > 3 {
			return nil, dimXY, errInvalidKML
		}
		if len(point) != len(points[0]) {
			return nil, dimXY, pgerror.New(pgcode.Internal, "invalid KML representation: mixed coordinates dimension")
		}
		seq = append(seq, point...)
	}
	if len(points[0]) == 3 {
		return seq, dimXYZ, nil
	}
	return seq, dimXY, nil
}

// kmlRoutines returns the routines that read KML.
func kmlRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_geomfromkml", Symbol: "geom_from_kml", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromKML},
	}
}
