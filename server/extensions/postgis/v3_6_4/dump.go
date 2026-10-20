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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// extensionType returns the named type that the extension declares, from the schema that the extension is installed
// in.
func extensionType(ctx *sql.Context, typeName string) (*pgtypes.DoltgresType, error) {
	extColl, err := core.GetExtensionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	ext, err := extColl.GetLoadedExtension(ctx, id.NewExtension("postgis"))
	if err != nil {
		return nil, err
	}
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	typ, err := typColl.GetType(ctx, id.NewType(ext.Namespace.SchemaName(), typeName))
	if err != nil || typ == nil {
		return nil, pgtypes.ErrTypeDoesNotExist.New(typeName)
	}
	return typ, nil
}

// compositeValue returns a value of the named composite type that the extension declares, holding the given
// attribute values in order.
func compositeValue(ctx *sql.Context, typeName string, values ...any) ([]pgtypes.RecordValue, error) {
	typ, err := extensionType(ctx, typeName)
	if err != nil {
		return nil, err
	}
	record := make([]pgtypes.RecordValue, len(values))
	for i, value := range values {
		record[i] = pgtypes.RecordValue{Value: value, Type: typ.CompositeAttrs[i].Type}
	}
	return record, nil
}

// dumpRow is one row of the dump functions: the path of positions that leads to a part of a geometry, and the part.
type dumpRow struct {
	path []int32
	part shape
}

// dumpIter returns the rows of geometry_dump values for the parts of a geometry, which keep its SRID.
func dumpIter(ctx *sql.Context, g Geometry, rows []dumpRow) (sql.RowIter, error) {
	sqlRows := make([]sql.Row, len(rows))
	for i, row := range rows {
		path := make([]any, len(row.path))
		for j, position := range row.path {
			path[j] = position
		}
		value, err := compositeValue(ctx, "geometry_dump", path, g.withShape(row.part))
		if err != nil {
			return nil, err
		}
		sqlRows[i] = sql.Row{value}
	}
	return sql.RowsToRowIter(sqlRows...), nil
}

// extendPath returns a copy of the path with more positions after it.
func extendPath(path []int32, positions ...int32) []int32 {
	return append(append([]int32(nil), path...), positions...)
}

// hasMembers returns whether shapes of the type number hold their parts as members, which the dump functions descend
// into.
func hasMembers(typeNum int) bool {
	return !hasSequence(typeNum) && typeNum != polygonType
}

// dumpParts returns the rows of ST_Dump: every part of the shape that has no members, including empty parts inside
// collections, with the 1-based positions that lead to it.
func dumpParts(s shape, path []int32) []dumpRow {
	if !hasMembers(s.typeNum) {
		return []dumpRow{{path: path, part: s}}
	}
	var rows []dumpRow
	for i, member := range s.members {
		rows = append(rows, dumpParts(member, extendPath(path, int32(i+1)))...)
	}
	return rows
}

// stDump implements ST_Dump. A single shape dumps itself with an empty path, unless it is empty.
func stDump(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if !hasMembers(g.Shape.typeNum) && g.Shape.isEmpty() {
		return sql.RowsToRowIter(), nil
	}
	return dumpIter(ctx, g, dumpParts(g.Shape, []int32{}))
}

// dumpPoints returns the rows of ST_DumpPoints: every vertex of the shape, with the positions of its part, its ring
// for a surface, and its position in the sequence.
func dumpPoints(s shape, path []int32) []dumpRow {
	dims := s.ctype.Dimension()
	pointsOf := func(seq []float64, prefix []int32) []dumpRow {
		var rows []dumpRow
		for i := 0; i+dims <= len(seq); i += dims {
			rows = append(rows, dumpRow{path: extendPath(prefix, int32(i/dims+1)), part: pointShape(seq, s.ctype, i/dims)})
		}
		return rows
	}
	switch s.typeNum {
	case pointType:
		if s.isEmpty() {
			return nil
		}
		return []dumpRow{{path: extendPath(path, 1), part: s}}
	case lineType, circStringType:
		return pointsOf(s.coords, path)
	case triangleType:
		return pointsOf(s.coords, extendPath(path, 1))
	case polygonType:
		var rows []dumpRow
		for i, ring := range s.rings {
			rows = append(rows, pointsOf(ring, extendPath(path, int32(i+1)))...)
		}
		return rows
	default:
		var rows []dumpRow
		for i, member := range s.members {
			rows = append(rows, dumpPoints(member, extendPath(path, int32(i+1)))...)
		}
		return rows
	}
}

// stDumpPoints implements ST_DumpPoints.
func stDumpPoints(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	return dumpIter(ctx, g, dumpPoints(g.Shape, []int32{}))
}

// stDumpRings implements ST_DumpRings, which dumps each ring of a polygon as a polygon, numbering the exterior ring
// zero.
func stDumpRings(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if g.Shape.typeNum != polygonType {
		return nil, pgerror.New(pgcode.Internal, "Input is not a polygon")
	}
	rows := make([]dumpRow, len(g.Shape.rings))
	for i, ring := range g.Shape.rings {
		rows[i] = dumpRow{path: []int32{int32(i)}, part: shape{typeNum: polygonType, ctype: g.Shape.ctype, rings: [][]float64{ring}}}
	}
	return dumpIter(ctx, g, rows)
}

// dumpSegments returns the rows of ST_DumpSegments: every segment of the shape's linestrings and rings as a
// linestring, and every arc of its circular strings as a circular string.
func dumpSegments(s shape, path []int32) []dumpRow {
	dims := s.ctype.Dimension()
	segmentsOf := func(seq []float64, prefix []int32, typeNum int) []dumpRow {
		step := 1
		if typeNum == circStringType {
			step = 2
		}
		var rows []dumpRow
		for i := 0; i+step*dims < len(seq); i += step * dims {
			part := shape{typeNum: typeNum, ctype: s.ctype, coords: append([]float64(nil), seq[i:i+(step+1)*dims]...)}
			rows = append(rows, dumpRow{path: extendPath(prefix, int32(len(rows)+1)), part: part})
		}
		return rows
	}
	switch s.typeNum {
	case pointType:
		return nil
	case lineType:
		return segmentsOf(s.coords, path, lineType)
	case circStringType:
		return segmentsOf(s.coords, path, circStringType)
	case triangleType:
		return segmentsOf(s.coords, extendPath(path, 1), lineType)
	case polygonType:
		var rows []dumpRow
		for i, ring := range s.rings {
			rows = append(rows, segmentsOf(ring, extendPath(path, int32(i+1)), lineType)...)
		}
		return rows
	default:
		var rows []dumpRow
		for i, member := range s.members {
			rows = append(rows, dumpSegments(member, extendPath(path, int32(i+1)))...)
		}
		return rows
	}
}

// stDumpSegments implements ST_DumpSegments.
func stDumpSegments(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	return dumpIter(ctx, g, dumpSegments(g.Shape, []int32{}))
}

// composites returns the composite types that the extension declares.
func composites() []extdef.Composite {
	return []extdef.Composite{
		{Name: "geometry_dump", Attributes: []extdef.Parameter{{Name: "path", Type: "_int4"}, {Name: "geom", Type: "geometry"}}},
		{Name: "valid_detail", Attributes: []extdef.Parameter{{Name: "valid", Type: "bool"}, {Name: "reason", Type: "varchar"}, {Name: "location", Type: "geometry"}}},
	}
}

// dumpRoutines returns the routines that dump the parts of geometries.
func dumpRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_dump", Symbol: "LWGEOM_dump", Parameters: params("geometry"), Returns: "geometry_dump", Strict: true, SetOf: true, Impl: stDump},
		{Name: "st_dumppoints", Symbol: "LWGEOM_dumppoints", Parameters: params("geometry"), Returns: "geometry_dump", Strict: true, SetOf: true, Impl: stDumpPoints},
		{Name: "st_dumprings", Symbol: "LWGEOM_dump_rings", Parameters: params("geometry"), Returns: "geometry_dump", Strict: true, SetOf: true, Impl: stDumpRings},
		{Name: "st_dumpsegments", Symbol: "LWGEOM_dumpsegments", Parameters: params("geometry"), Returns: "geometry_dump", Strict: true, SetOf: true, Impl: stDumpSegments},
	}
}
