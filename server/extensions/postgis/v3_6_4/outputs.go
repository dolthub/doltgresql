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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// precisionArg returns the maximum number of decimal digits passed as the given argument, or the default when the
// call omitted it.
func precisionArg(args []any, index int) int {
	if len(args) > index {
		return int(args[index].(int32))
	}
	return defaultPrecision
}

// byteOrderArg returns the byte order named by the given argument, which is big-endian only for "XDR" in any case.
func byteOrderArg(args []any, index int) binary.AppendByteOrder {
	if len(args) > index && strings.EqualFold(args[index].(string), "XDR") {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// stAsText implements ST_AsText.
func stAsText(ctx *sql.Context, args ...any) (any, error) {
	return writeWKT(args[0].(Geometry), precisionArg(args, 1), wktISO), nil
}

// stAsEWKT implements ST_AsEWKT.
func stAsEWKT(ctx *sql.Context, args ...any) (any, error) {
	return writeWKT(args[0].(Geometry), precisionArg(args, 1), wktExtended), nil
}

// stAsBinary implements ST_AsBinary, which writes ISO WKB.
func stAsBinary(ctx *sql.Context, args ...any) (any, error) {
	return encodeWKB(args[0].(Geometry), byteOrderArg(args, 1), wkbISO), nil
}

// stAsEWKB implements ST_AsEWKB.
func stAsEWKB(ctx *sql.Context, args ...any) (any, error) {
	return encodeWKB(args[0].(Geometry), byteOrderArg(args, 1), wkbExtended), nil
}

// stAsHexEWKB implements ST_AsHEXEWKB.
func stAsHexEWKB(ctx *sql.Context, args ...any) (any, error) {
	return strings.ToUpper(hex.EncodeToString(encodeWKB(args[0].(Geometry), byteOrderArg(args, 1), wkbExtended))), nil
}

// stAsBinaryNullable implements ST_AsBinary(geography, text), which is not strict and treats a null byte order as
// little-endian.
func stAsBinaryNullable(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil {
		return nil, nil
	}
	if args[1] == nil {
		return stAsBinary(ctx, args[0])
	}
	return stAsBinary(ctx, args...)
}

// outputRoutines returns the routines that write geometries and geographies in other formats.
func outputRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_asbinary", Symbol: "st_asbinary_geography", Parameters: params("geography"), Returns: "bytea", Strict: true, Impl: stAsBinary},
		{Name: "st_asbinary", Symbol: "st_asbinary_geography_text", Parameters: params("geography", "text"), Returns: "bytea", Impl: stAsBinaryNullable},
		{Name: "st_asbinary", Symbol: "st_asbinary_geometry", Parameters: params("geometry"), Returns: "bytea", Strict: true, Impl: stAsBinary},
		{Name: "st_asbinary", Symbol: "st_asbinary_geometry_text", Parameters: params("geometry", "text"), Returns: "bytea", Strict: true, Impl: stAsBinary},
		{Name: "st_asewkb", Symbol: "st_asewkb_geometry", Parameters: params("geometry"), Returns: "bytea", Strict: true, Impl: stAsEWKB},
		{Name: "st_asewkb", Symbol: "st_asewkb_geometry_text", Parameters: params("geometry", "text"), Returns: "bytea", Strict: true, Impl: stAsEWKB},
		{Name: "st_asewkt", Symbol: "st_asewkt_geography", Parameters: params("geography"), Returns: "text", Strict: true, Impl: stAsEWKT},
		{Name: "st_asewkt", Symbol: "st_asewkt_geography_int4", Parameters: params("geography", "int4"), Returns: "text", Strict: true, Impl: stAsEWKT},
		{Name: "st_asewkt", Symbol: "st_asewkt_geometry", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: stAsEWKT},
		{Name: "st_asewkt", Symbol: "st_asewkt_geometry_int4", Parameters: params("geometry", "int4"), Returns: "text", Strict: true, Impl: stAsEWKT},
		{Name: "st_asewkt", Symbol: "st_asewkt_text", Parameters: params("text"), Returns: "text", Strict: true, Impl: textArgsAsGeometry(1, stAsEWKT)},
		{Name: "st_ashexewkb", Symbol: "st_ashexewkb_geometry", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: stAsHexEWKB},
		{Name: "st_ashexewkb", Symbol: "st_ashexewkb_geometry_text", Parameters: params("geometry", "text"), Returns: "text", Strict: true, Impl: stAsHexEWKB},
		{Name: "st_astext", Symbol: "st_astext_geography", Parameters: params("geography"), Returns: "text", Strict: true, Impl: stAsText},
		{Name: "st_astext", Symbol: "st_astext_geography_int4", Parameters: params("geography", "int4"), Returns: "text", Strict: true, Impl: stAsText},
		{Name: "st_astext", Symbol: "st_astext_geometry", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: stAsText},
		{Name: "st_astext", Symbol: "st_astext_geometry_int4", Parameters: params("geometry", "int4"), Returns: "text", Strict: true, Impl: stAsText},
		{Name: "st_astext", Symbol: "st_astext_text", Parameters: params("text"), Returns: "text", Strict: true, Impl: textArgsAsGeometry(1, stAsText)},
	}
}
