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
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// listedCRSs returns the coordinate reference systems of an authority that the postgis_srs functions list: the
// geographic, projected, and compound systems that are not deprecated, each code once. The projected systems come
// first, then the geographic systems, and then the compound systems, each in the text order of their codes.
func listedCRSs(authority string) ([]spatial.CRSInfo, error) {
	infos, err := spatial.DatabaseCRSInfo(authority, nil)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var listed []spatial.CRSInfo
	for _, info := range infos {
		switch info.Kind {
		case spatial.Geographic2DCRS, spatial.Geographic3DCRS, spatial.ProjectedCRS, spatial.CompoundCRS:
		default:
			continue
		}
		if info.Deprecated || seen[info.Code] {
			continue
		}
		seen[info.Code] = true
		listed = append(listed, info)
	}
	group := func(info spatial.CRSInfo) int {
		switch info.Kind {
		case spatial.ProjectedCRS:
			return 0
		case spatial.CompoundCRS:
			return 2
		default:
			return 1
		}
	}
	sort.Slice(listed, func(i, j int) bool {
		if group(listed[i]) != group(listed[j]) {
			return group(listed[i]) < group(listed[j])
		}
		return listed[i].Code < listed[j].Code
	})
	return listed, nil
}

// srsRows returns the rows of the postgis_srs functions for the given systems: each system's authority, code, name,
// WKT, PROJ string, and the corners of its area of use.
func srsRows(ctx *sql.Context, infos []spatial.CRSInfo) (sql.RowIter, error) {
	geometryType, err := extensionType(ctx, "geometry")
	if err != nil {
		return nil, err
	}
	rows := make([]sql.Row, 0, len(infos))
	for _, info := range infos {
		wkt, projString, err := spatial.CRSDefinitions(info.Authority, info.Code)
		if err != nil {
			return nil, err
		}
		var southWest, northEast any
		if info.AreaValid {
			southWest = Geometry{SRID: sridDefault, Shape: shape{typeNum: pointType, coords: []float64{info.West, info.South}}}
			northEast = Geometry{SRID: sridDefault, Shape: shape{typeNum: pointType, coords: []float64{info.East, info.North}}}
		}
		rows = append(rows, sql.Row{[]pgtypes.RecordValue{
			{Value: info.Authority, Type: pgtypes.Text},
			{Value: info.Code, Type: pgtypes.Text},
			{Value: info.Name, Type: pgtypes.Text},
			{Value: wkt, Type: pgtypes.Text},
			{Value: projString, Type: pgtypes.Text},
			{Value: southWest, Type: geometryType},
			{Value: northEast, Type: geometryType},
		}})
	}
	return sql.RowsToRowIter(rows...), nil
}

// postgisSRSCodes implements postgis_srs_codes, which returns the codes of an authority's listed systems.
func postgisSRSCodes(ctx *sql.Context, args ...any) (any, error) {
	infos, err := listedCRSs(args[0].(string))
	if err != nil {
		return nil, err
	}
	rows := make([]sql.Row, len(infos))
	for i, info := range infos {
		rows[i] = sql.Row{info.Code}
	}
	return sql.RowsToRowIter(rows...), nil
}

// postgisSRS implements postgis_srs, which returns the row of one system, or a row of nulls when PROJ does not know
// it.
func postgisSRS(ctx *sql.Context, args ...any) (any, error) {
	authority, code := args[0].(string), args[1].(string)
	infos, _ := spatial.DatabaseCRSInfo(authority, nil)
	for _, info := range infos {
		if info.Code == code {
			return srsRows(ctx, []spatial.CRSInfo{info})
		}
	}
	geometryType, err := extensionType(ctx, "geometry")
	if err != nil {
		return nil, err
	}
	nulls := make([]pgtypes.RecordValue, 7)
	for i := range nulls {
		nulls[i].Type = pgtypes.Text
	}
	nulls[5].Type, nulls[6].Type = geometryType, geometryType
	return sql.RowsToRowIter(sql.Row{nulls}), nil
}

// postgisSRSAll implements postgis_srs_all, which returns the rows of the listed systems of EPSG, ESRI, and the IAU.
func postgisSRSAll(ctx *sql.Context, args ...any) (any, error) {
	var all []spatial.CRSInfo
	for _, authority := range []string{"EPSG", "ESRI", "IAU_2015"} {
		infos, err := listedCRSs(authority)
		if err != nil {
			return nil, err
		}
		all = append(all, infos...)
	}
	return srsRows(ctx, all)
}

// postgisSRSSearch implements postgis_srs_search, which returns the rows of an authority's projected systems whose
// areas of use hold the bounding box of the geometry in WGS 84, from the smallest area to the largest, and by code
// among equal areas.
func postgisSRSSearch(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	bounds := args[0].(Geometry)
	if bounds.SRID != sridDefault {
		if _, err := lookUpSRS(ctx, runner, bounds.SRID); err != nil {
			return nil, err
		}
		transformed, err := stTransform(ctx, runner, bounds, int32(sridDefault))
		if err != nil {
			return nil, err
		}
		bounds = transformed.(Geometry)
	}
	b, ok := bounds.Shape.bounds()
	if !ok {
		return sql.RowsToRowIter(), nil
	}
	infos, err := spatial.DatabaseCRSInfo(args[1].(string), &[4]float64{b.minX, b.minY, b.maxX, b.maxY})
	if err != nil {
		return nil, err
	}
	area := func(info spatial.CRSInfo) float64 {
		width := info.East - info.West
		if width < 0 {
			width += 360
		}
		return width * (info.North - info.South)
	}
	sort.SliceStable(infos, func(i, j int) bool {
		if area(infos[i]) != area(infos[j]) {
			return area(infos[i]) < area(infos[j])
		}
		return infos[i].Code < infos[j].Code
	})
	return srsRows(ctx, infos)
}

// srsRoutines returns the routines that list the coordinate reference systems in PROJ's database.
func srsRoutines() []extdef.Routine {
	srsOutputs := []extdef.Parameter{{Name: "auth_name", Type: "text"}, {Name: "auth_srid", Type: "text"}, {Name: "srname", Type: "text"}, {Name: "srtext", Type: "text"}, {Name: "proj4text", Type: "text"}, {Name: "point_sw", Type: "geometry"}, {Name: "point_ne", Type: "geometry"}}
	return []extdef.Routine{
		{Name: "postgis_srs", Symbol: "postgis_srs_entry", Parameters: []extdef.Parameter{{Name: "auth_name", Type: "text"}, {Name: "auth_srid", Type: "text"}}, Outputs: srsOutputs, Strict: true, SetOf: true, Impl: postgisSRS},
		{Name: "postgis_srs_all", Symbol: "postgis_srs_entry_all", Parameters: params(), Outputs: srsOutputs, Strict: true, SetOf: true, Impl: postgisSRSAll},
		{Name: "postgis_srs_codes", Symbol: "postgis_srs_codes", Parameters: []extdef.Parameter{{Name: "auth_name", Type: "text"}}, Returns: "text", Strict: true, SetOf: true, Impl: postgisSRSCodes},
		{Name: "postgis_srs_search", Symbol: "postgis_srs_search", Parameters: []extdef.Parameter{{Name: "bounds", Type: "geometry"}, {Name: "authname", Type: "text", Default: `'EPSG'::text`}}, Outputs: srsOutputs, Strict: true, SetOf: true, QueryImpl: postgisSRSSearch},
	}
}
