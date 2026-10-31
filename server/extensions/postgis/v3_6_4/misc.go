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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// postgisVersion is the version of PostGIS that the extension emulates.
const postgisVersion = "3.6.4"

// postgisRevision is the revision of the PostGIS source that the emulated version was released from.
const postgisRevision = "94d984b"

// postgisRevisionDate is when the emulated revision was committed, which stands in for the build date of the library
// and its scripts.
const postgisRevisionDate = "2026-06-08 18:28:56"

// wagyuVersion is the version of the Wagyu library in third_party/spatial, which PostGIS reports as built in.
const wagyuVersion = "0.5.0 (Internal)"

// constant returns an implementation that always returns the value.
func constant(value any) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return value, nil
	}
}

// stPointWithDims returns an implementation of ST_PointZ, ST_PointM, or ST_PointZM, which take their ordinates and
// then an SRID.
func stPointWithDims(ctype coordinatesType) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		dims := ctype.Dimension()
		coords := floatArgs(args[:dims], 0)
		return Geometry{SRID: clampSRID(ctx, args[dims].(int32)), Shape: shape{typeNum: pointType, ctype: ctype, coords: coords}}, nil
	}
}

// stMakePointM implements ST_MakePointM.
func stMakePointM(ctx *sql.Context, args ...any) (any, error) {
	return Geometry{Shape: shape{typeNum: pointType, ctype: dimXYM, coords: floatArgs(args, 0)}}, nil
}

// postgisTypeName implements postgis_type_name, which names a geometry type with the given number of dimensions:
// with a Z, M, or ZM suffix in the new style, and in upper case with only an M suffix in the old style. An unknown
// type or dimension count returns null.
func postgisTypeName(ctx *sql.Context, args ...any) (any, error) {
	name, dims, newStyle := strings.ToUpper(args[0].(string)), args[1].(int32), args[2].(bool)
	typeNum, ok := typmodTypeNumber(name)
	measured := false
	if !ok && strings.HasSuffix(name, "M") {
		typeNum, ok = typmodTypeNumber(strings.TrimSuffix(name, "M"))
		measured = ok
	}
	if !ok || dims < 2 || dims > 4 || (measured && dims != 3) {
		return nil, nil
	}
	if !newStyle {
		return name, nil
	}
	t := makeTypmod(0, typeNum, dims == 4 || (dims == 3 && !measured), dims == 4 || measured)
	return t.typeName(), nil
}

// postgisDeprecate implements _postgis_deprecate, which warns that a signature is deprecated when it was deprecated
// in an earlier major or minor version than the emulated one.
func postgisDeprecate(ctx *sql.Context, args ...any) (any, error) {
	oldName, newName, version := args[0].(string), args[1].(string), args[2].(string)
	current := strings.Split(postgisVersion, ".")
	deprecated := strings.Split(version, ".")
	if len(deprecated) > 1 && (current[0] != deprecated[0] || current[1] != deprecated[1]) {
		sendNotice(ctx, "WARNING", fmt.Sprintf("%s signature was deprecated in %s. Please use %s", oldName, version, newName))
	}
	return "", nil
}

// miscRoutines returns the version, type modifier, and bounding box helper routines.
func miscRoutines() []extdef.Routine {
	geometryIdentity := func(ctx *sql.Context, args ...any) (any, error) {
		return args[0], nil
	}
	withSRID := func(names ...string) []extdef.Parameter {
		ps := make([]extdef.Parameter, len(names))
		for i, name := range names {
			ps[i] = extdef.Parameter{Name: name, Type: "float8"}
		}
		return append(ps, extdef.Parameter{Name: "srid", Type: "int4", Default: `0`})
	}
	return []extdef.Routine{
		{Name: "postgis_addbbox", Symbol: "LWGEOM_addBBOX", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: geometryIdentity},
		{Name: "postgis_dropbbox", Symbol: "LWGEOM_dropBBOX", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: geometryIdentity},
		{Name: "postgis_full_version", Symbol: "postgis_full_version", Parameters: params(), Returns: "text", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return `POSTGIS="` + postgisVersion + ` ` + postgisRevision + `" [EXTENSION] PGSQL="150" GEOS="` + spatial.GEOSVersion() + `" PROJ="` + projVersion() + `" WAGYU="` + wagyuVersion + `"`, nil
		}},
		{Name: "postgis_geos_compiled_version", Symbol: "postgis_geos_compiled_version", Parameters: params(), Returns: "text", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			version, _, _ := strings.Cut(spatial.GEOSVersion(), "-")
			return version, nil
		}},
		{Name: "postgis_geos_version", Symbol: "postgis_geos_version", Parameters: params(), Returns: "text", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return spatial.GEOSVersion(), nil
		}},
		{Name: "postgis_geos_noop", Symbol: "GEOSnoop", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: geometryIdentity},
		{Name: "postgis_getbbox", Symbol: "LWGEOM_to_BOX2DF", Parameters: params("geometry"), Returns: "box2d", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			b, ok := args[0].(Geometry).Shape.bounds()
			if !ok {
				return nil, nil
			}
			b = b.floatRounded()
			return box2d{minX: b.minX, minY: b.minY, maxX: b.maxX, maxY: b.maxY}, nil
		}},
		{Name: "postgis_hasbbox", Symbol: "LWGEOM_hasBBOX", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			s := args[0].(Geometry).Shape
			return !s.isEmpty() && !hasExactBox(s), nil
		}},
		{Name: "postgis_lib_build_date", Symbol: "postgis_lib_build_date", Parameters: params(), Returns: "text", Impl: constant(postgisRevisionDate)},
		{Name: "postgis_lib_version", Symbol: "postgis_lib_version", Parameters: params(), Returns: "text", Impl: constant(postgisVersion)},
		{Name: "postgis_liblwgeom_version", Symbol: "postgis_liblwgeom_version", Parameters: params(), Returns: "text", Impl: constant(postgisVersion)},
		{Name: "postgis_libjson_version", Symbol: "postgis_libjson_version", Parameters: params(), Returns: "text", Impl: constant(nil)},
		{Name: "postgis_libprotobuf_version", Symbol: "postgis_libprotobuf_version", Parameters: params(), Returns: "text", Impl: constant(nil)},
		{Name: "postgis_libxml_version", Symbol: "postgis_libxml_version", Parameters: params(), Returns: "text", Impl: constant(nil)},
		{Name: "postgis_noop", Symbol: "LWGEOM_noop", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: geometryIdentity},
		{Name: "postgis_proj_compiled_version", Symbol: "postgis_proj_compiled_version", Parameters: params(), Returns: "text", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return spatial.PROJVersion(), nil
		}},
		{Name: "postgis_proj_version", Symbol: "postgis_proj_version", Parameters: params(), Returns: "text", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return projVersion(), nil
		}},
		{Name: "postgis_scripts_build_date", Symbol: "postgis_scripts_build_date", Parameters: params(), Returns: "text", Impl: constant(postgisRevisionDate)},
		{Name: "postgis_scripts_installed", Symbol: "postgis_scripts_installed", Parameters: params(), Returns: "text", Impl: constant(postgisVersion + " " + postgisRevision)},
		{Name: "postgis_scripts_released", Symbol: "postgis_scripts_released", Parameters: params(), Returns: "text", Impl: constant(postgisVersion + " " + postgisRevision)},
		{Name: "postgis_lib_revision", Symbol: "postgis_lib_revision", Parameters: params(), Returns: "text", Impl: constant(postgisRevision)},
		{Name: "postgis_svn_version", Symbol: "postgis_svn_version", Parameters: params(), Returns: "text", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			if _, err := postgisDeprecate(ctx, "postgis_svn_version", "postgis_lib_revision", "3.1.0"); err != nil {
				return nil, err
			}
			return postgisRevision, nil
		}},
		{Name: "_postgis_deprecate", Symbol: "_postgis_deprecate", Parameters: []extdef.Parameter{{Name: "oldname", Type: "text"}, {Name: "newname", Type: "text"}, {Name: "version", Type: "text"}}, Returns: "void", Strict: true, Impl: postgisDeprecate},
		{Name: "_postgis_pgsql_version", Symbol: "_postgis_pgsql_version", Parameters: params(), Returns: "text", Impl: constant("150")},
		{Name: "_postgis_scripts_pgsql_version", Symbol: "_postgis_scripts_pgsql_version", Parameters: params(), Returns: "text", Impl: constant("150")},
		{Name: "postgis_type_name", Symbol: "postgis_type_name_varchar_int4_bool", Parameters: []extdef.Parameter{{Name: "geomname", Type: "varchar"}, {Name: "coord_dimension", Type: "int4"}, {Name: "use_new_name", Type: "bool", Default: `true`}}, Returns: "varchar", Strict: true, Impl: postgisTypeName},
		{Name: "postgis_typmod_dims", Symbol: "postgis_typmod_dims", Parameters: params("int4"), Returns: "int4", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			t := typmod(args[0].(int32))
			if t < 0 {
				return nil, nil
			}
			dims := int32(2)
			if t.hasZ() {
				dims++
			}
			if t.hasM() {
				dims++
			}
			return dims, nil
		}},
		{Name: "postgis_typmod_srid", Symbol: "postgis_typmod_srid", Parameters: params("int4"), Returns: "int4", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			if t := typmod(args[0].(int32)); t >= 0 {
				return t.srid(), nil
			}
			return int32(0), nil
		}},
		{Name: "postgis_typmod_type", Symbol: "postgis_typmod_type", Parameters: params("int4"), Returns: "text", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return typmod(args[0].(int32)).typeName(), nil
		}},
		{Name: "postgis_version", Symbol: "postgis_version", Parameters: params(), Returns: "text", Impl: constant("3.6 USE_GEOS=1 USE_PROJ=1 USE_STATS=1")},
		{Name: "postgis_wagyu_version", Symbol: "postgis_wagyu_version", Parameters: params(), Returns: "text", Impl: constant(wagyuVersion)},
		{Name: "st_makepointm", Symbol: "LWGEOM_makepoint3dm", Parameters: params("float8", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stMakePointM},
		{Name: "st_pointm", Symbol: "ST_PointM", Parameters: withSRID("xcoordinate", "ycoordinate", "mcoordinate"), Returns: "geometry", Strict: true, Impl: stPointWithDims(dimXYM)},
		{Name: "st_pointz", Symbol: "ST_PointZ", Parameters: withSRID("xcoordinate", "ycoordinate", "zcoordinate"), Returns: "geometry", Strict: true, Impl: stPointWithDims(dimXYZ)},
		{Name: "st_pointzm", Symbol: "ST_PointZM", Parameters: withSRID("xcoordinate", "ycoordinate", "zcoordinate", "mcoordinate"), Returns: "geometry", Strict: true, Impl: stPointWithDims(dimXYZM)},
	}
}

// projVersion returns the version of PROJ along with its network and data settings, as postgis_proj_version reports
// them.
func projVersion() string {
	networkEnabled, urlEndpoint, writableDirectory, databasePath := spatial.PROJSettings()
	network := "OFF"
	if networkEnabled {
		network = "ON"
	}
	return spatial.PROJVersion() + " NETWORK_ENABLED=" + network + " URL_ENDPOINT=" + urlEndpoint +
		" USER_WRITABLE_DIRECTORY=" + writableDirectory + " DATABASE_PATH=" + databasePath
}
