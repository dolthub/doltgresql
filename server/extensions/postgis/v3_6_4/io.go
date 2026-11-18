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
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// parseGeometryInput parses the text forms that the geometry type accepts: hex-encoded WKB, GeoJSON, or WKT, each of
// which may be preceded by an "SRID=<srid>;" prefix.
func parseGeometryInput(ctx *sql.Context, input string) (Geometry, error) {
	if input == "" {
		return Geometry{}, errWKTInvalid
	}
	body := input
	var srid int32
	if len(input) >= 5 && strings.EqualFold(input[:5], "SRID=") {
		if end := strings.IndexByte(input, ';'); end >= 0 && end+1 < len(input) && input[end+1] == '0' {
			parsed, _ := strconv.Atoi(input[5:end])
			srid, body = int32(parsed), input[end+1:]
		}
	}
	switch body[0] {
	case '0':
		g, err := decodeHexWKB(ctx, body)
		if err != nil {
			return Geometry{}, err
		}
		if srid != sridUnknown {
			g.SRID = clampSRID(ctx, srid)
		}
		return g, nil
	case '{':
		g, _, err := parseGeoJSON(ctx, body)
		return g, err
	default:
		g, err := parseEWKT(body)
		if err != nil {
			return Geometry{}, err
		}
		g.SRID = clampSRID(ctx, g.SRID)
		return g, nil
	}
}

// decodeHexWKB decodes hex-encoded extended or ISO WKB.
func decodeHexWKB(ctx *sql.Context, input string) (Geometry, error) {
	if len(input)%2 != 0 {
		return Geometry{}, pgerror.Newf(pgcode.Internal, "Invalid hex string, length (%d) has to be a multiple of two!", len(input))
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		return Geometry{}, pgerror.New(pgcode.Internal, "Invalid hex string")
	}
	return decodeWKB(ctx, data)
}

// encodeHexWKB returns the upper-case hex encoding of the geometry's little-endian extended WKB.
func encodeHexWKB(g Geometry) string {
	return strings.ToUpper(hex.EncodeToString(encodeWKB(g, binary.LittleEndian, wkbExtended)))
}

// typmodArg returns the type modifier passed as the third argument of an input or receive function, or -1 when the
// caller passed none.
func typmodArg(args []any) typmod {
	if len(args) > 2 && args[2] != nil {
		return typmod(args[2].(int32))
	}
	return -1
}

// geometryIn implements geometry_in.
func geometryIn(ctx *sql.Context, args ...any) (any, error) {
	g, err := parseGeometryInput(ctx, args[0].(string))
	if err != nil {
		return nil, err
	}
	return enforceTypmod(g, typmodArg(args))
}

// geometryOut implements geometry_out.
func geometryOut(ctx *sql.Context, args ...any) (any, error) {
	return encodeHexWKB(args[0].(Geometry)), nil
}

// geometryRecv implements geometry_recv.
func geometryRecv(ctx *sql.Context, args ...any) (any, error) {
	g, err := decodeWKB(ctx, args[0].([]byte))
	if err != nil {
		return nil, err
	}
	return enforceTypmod(g, typmodArg(args))
}

// geometrySend implements geometry_send.
func geometrySend(ctx *sql.Context, args ...any) (any, error) {
	return encodeWKB(args[0].(Geometry), binary.LittleEndian, wkbExtended), nil
}

// geometryTypmodIn implements geometry_typmod_in.
func geometryTypmodIn(ctx *sql.Context, args ...any) (any, error) {
	t, err := parseTypmod(ctx, args[0].([]any), false)
	return int32(t), err
}

// geographyTypmodIn implements geography_typmod_in.
func geographyTypmodIn(ctx *sql.Context, args ...any) (any, error) {
	t, err := parseTypmod(ctx, args[0].([]any), true)
	return int32(t), err
}

// postgisTypmodOut implements postgis_typmod_out, which both types use.
func postgisTypmodOut(ctx *sql.Context, args ...any) (any, error) {
	return formatTypmod(typmod(args[0].(int32))), nil
}

// geometryEnforceTypmod implements geometry_enforce_typmod, the cast that applies a type modifier.
func geometryEnforceTypmod(ctx *sql.Context, args ...any) (any, error) {
	return enforceTypmod(args[0].(Geometry), typmod(args[1].(int32)))
}

// geometryFromText implements geometry(text), the cast from text to geometry, which accepts the same forms as
// geometry_in.
func geometryFromText(ctx *sql.Context, args ...any) (any, error) {
	return parseGeometryInput(ctx, args[0].(string))
}

// geometryFromBytea implements geometry(bytea) and ST_GeomFromEWKB, which decode extended WKB.
func geometryFromBytea(ctx *sql.Context, args ...any) (any, error) {
	return decodeWKB(ctx, args[0].([]byte))
}

// geographyFromBytea implements geography(bytea) and ST_GeogFromWKB, which decode WKB into a geography.
func geographyFromBytea(ctx *sql.Context, args ...any) (any, error) {
	g, err := decodeWKB(ctx, args[0].([]byte))
	if err != nil {
		return nil, err
	}
	return asGeography(ctx, g)
}

// ioRoutines returns the input, output, type modifier, and cast routines of the geometry and geography types.
func ioRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "geometry_in", Symbol: "LWGEOM_in", Parameters: params("cstring"), Returns: "geometry", Strict: true, Impl: geometryIn},
		{Name: "geometry_out", Symbol: "LWGEOM_out", Parameters: params("geometry"), Returns: "cstring", Strict: true, Impl: geometryOut},
		{Name: "geometry_recv", Symbol: "LWGEOM_recv", Parameters: params("internal"), Returns: "geometry", Strict: true, Impl: geometryRecv},
		{Name: "geometry_send", Symbol: "LWGEOM_send", Parameters: params("geometry"), Returns: "bytea", Strict: true, Impl: geometrySend},
		{Name: "geometry_typmod_in", Symbol: "geometry_typmod_in", Parameters: params("_cstring"), Returns: "int4", Strict: true, Impl: geometryTypmodIn},
		{Name: "geometry_typmod_out", Symbol: "geometry_typmod_out_int4", Parameters: params("int4"), Returns: "cstring", Strict: true, Impl: postgisTypmodOut},
		{Name: "geography_in", Symbol: "geography_in", Parameters: params("cstring", "oid", "int4"), Returns: "geography", Strict: true, Impl: geographyIn},
		{Name: "geography_out", Symbol: "geography_out", Parameters: params("geography"), Returns: "cstring", Strict: true, Impl: geometryOut},
		{Name: "geography_recv", Symbol: "geography_recv", Parameters: params("internal", "oid", "int4"), Returns: "geography", Strict: true, Impl: geographyRecv},
		{Name: "geography_send", Symbol: "geography_send", Parameters: params("geography"), Returns: "bytea", Strict: true, Impl: geometrySend},
		{Name: "geography_typmod_in", Symbol: "geography_typmod_in", Parameters: params("_cstring"), Returns: "int4", Strict: true, Impl: geographyTypmodIn},
		{Name: "geography_typmod_out", Symbol: "geography_typmod_out_int4", Parameters: params("int4"), Returns: "cstring", Strict: true, Impl: postgisTypmodOut},
		{Name: "geometry", Symbol: "geometry_enforce_typmod", Parameters: params("geometry", "int4", "bool"), Returns: "geometry", Strict: true, Impl: geometryEnforceTypmod},
		{Name: "geography", Symbol: "geography_enforce_typmod", Parameters: params("geography", "int4", "bool"), Returns: "geography", Strict: true, Impl: geometryEnforceTypmod},
		{Name: "geometry", Symbol: "geometry_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: geometryFromText},
		{Name: "text", Symbol: "LWGEOM_to_text", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: geometryOut},
		{Name: "geometry", Symbol: "LWGEOM_from_bytea", Parameters: params("bytea"), Returns: "geometry", Strict: true, Impl: geometryFromBytea},
		{Name: "bytea", Symbol: "bytea_geometry", Parameters: params("geometry"), Returns: "bytea", Strict: true, Impl: geometrySend},
		{Name: "geography", Symbol: "geography_bytea", Parameters: params("bytea"), Returns: "geography", Strict: true, Impl: geographyFromBytea},
		{Name: "bytea", Symbol: "bytea_geography", Parameters: params("geography"), Returns: "bytea", Strict: true, Impl: geometrySend},
		{Name: "geography", Symbol: "geography_from_geometry", Parameters: params("geometry"), Returns: "geography", Strict: true, Impl: geographyFromGeometry},
		{Name: "geometry", Symbol: "geometry_from_geography", Parameters: params("geography"), Returns: "geometry", Strict: true, Impl: geometryFromGeography},
	}
}

// ioCasts returns the casts between the geometry and geography types, to and from text and bytea, and to JSON.
func ioCasts() []extdef.Cast {
	return []extdef.Cast{
		{Source: "geometry", Target: "geometry", Routine: "geometry_enforce_typmod", CastType: casts.CastType_Implicit},
		{Source: "geography", Target: "geography", Routine: "geography_enforce_typmod", CastType: casts.CastType_Implicit},
		{Source: "text", Target: "geometry", Routine: "geometry_text", CastType: casts.CastType_Implicit},
		{Source: "geometry", Target: "text", Routine: "LWGEOM_to_text", CastType: casts.CastType_Implicit},
		{Source: "bytea", Target: "geometry", Routine: "LWGEOM_from_bytea", CastType: casts.CastType_Implicit},
		{Source: "geometry", Target: "bytea", Routine: "bytea_geometry", CastType: casts.CastType_Implicit},
		{Source: "bytea", Target: "geography", Routine: "geography_bytea", CastType: casts.CastType_Implicit},
		{Source: "geography", Target: "bytea", Routine: "bytea_geography", CastType: casts.CastType_Implicit},
		{Source: "geometry", Target: "geography", Routine: "geography_from_geometry", CastType: casts.CastType_Implicit},
		{Source: "geography", Target: "geometry", Routine: "geometry_from_geography", CastType: casts.CastType_Explicit},
		{Source: "geometry", Target: "json", Routine: "geometry_to_json", CastType: casts.CastType_Explicit},
		{Source: "geometry", Target: "jsonb", Routine: "geometry_to_jsonb", CastType: casts.CastType_Explicit},
	}
}
