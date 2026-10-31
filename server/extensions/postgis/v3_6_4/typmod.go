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
	"math"
	"strconv"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// typmod is a geometry or geography type modifier, which packs an SRID, a type number, and the Z and M flags into
// an integer.
type typmod int32

// srid returns the SRID that the type modifier requires, or zero when it requires none.
func (t typmod) srid() int32 {
	return ((int32(t) & 0x0FFFFF00) - (int32(t) & 0x10000000)) >> 8
}

// typeNum returns the type number that the type modifier requires, or zero when it allows any type.
func (t typmod) typeNum() int {
	return int((t & 0xFC) >> 2)
}

// hasZ returns whether the type modifier requires Z coordinates.
func (t typmod) hasZ() bool {
	return t&0x02 != 0
}

// hasM returns whether the type modifier requires M coordinates.
func (t typmod) hasM() bool {
	return t&0x01 != 0
}

// typeName returns the name of the type that the type modifier requires, with a Z, M, or ZM suffix for its dimensions,
// or "Geometry" when it allows any type.
func (t typmod) typeName() string {
	if t < 0 {
		return "Geometry"
	}
	name := "Geometry"
	if typeNum := t.typeNum(); typeNum != 0 {
		name = typeNameForNumber(typeNum)
	}
	if t.hasZ() {
		name += "Z"
	}
	if t.hasM() {
		name += "M"
	}
	return name
}

// makeTypmod packs the given SRID, type number, and dimension flags into a type modifier.
func makeTypmod(srid int32, typeNum int, z bool, m bool) typmod {
	t := typmod((srid&0x001FFFFF)<<8) | typmod((typeNum&0x3F)<<2)
	if z {
		t |= 0x02
	}
	if m {
		t |= 0x01
	}
	return t
}

// parseTypmod parses the strings of a geometry or geography type modifier: a type name with an optional Z, M, or ZM
// suffix, then an optional SRID. A geography's SRID defaults to 4326 and must be a longitude/latitude system.
func parseTypmod(ctx *sql.Context, mods []any, geography bool) (typmod, error) {
	var srid int32 = sridUnknown
	if geography {
		srid = sridDefault
	}
	typeNum := 0
	var z, m bool
	for i, mod := range mods {
		str := mod.(string)
		switch i {
		case 0:
			var ok bool
			if typeNum, z, m, ok = parseTypmodTypeName(str); !ok {
				return 0, pgerror.Newf(pgcode.InvalidParameterValue, "Invalid geometry type modifier: %s", str)
			}
		case 1:
			parsed, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
					return 0, pgerror.Newf(pgcode.NumericValueOutOfRange, `value "%s" is out of range for type integer`, str)
				}
				return 0, pgerror.Newf(pgcode.InvalidTextRepresentation, `invalid input syntax for type integer: "%s"`, str)
			}
			if parsed < math.MinInt32 || parsed > math.MaxInt32 {
				return 0, pgerror.Newf(pgcode.NumericValueOutOfRange, `value "%s" is out of range for type integer`, str)
			}
			if clamped := clampSRID(ctx, int32(parsed)); clamped != sridUnknown {
				srid = clamped
			}
		}
	}
	if geography {
		if err := checkLatLongSRID(srid); err != nil {
			return 0, err
		}
	}
	return makeTypmod(srid, typeNum, z, m), nil
}

// parseTypmodTypeName parses a type name with an optional Z, M, or ZM suffix, ignoring case.
func parseTypmodTypeName(str string) (typeNum int, z bool, m bool, ok bool) {
	name := strings.ToUpper(strings.TrimSpace(str))
	if typeNum, ok = typmodTypeNumber(name); ok {
		return typeNum, false, false, true
	}
	for _, suffix := range []string{"ZM", "Z", "M"} {
		if typeNum, ok = typmodTypeNumber(strings.TrimSuffix(name, suffix)); ok && strings.HasSuffix(name, suffix) {
			return typeNum, suffix != "M", suffix != "Z", true
		}
	}
	return 0, false, false, false
}

// typmodTypeNumber returns the type number of an upper-case type name in a type modifier, where "GEOMETRY" allows
// any type.
func typmodTypeNumber(name string) (int, bool) {
	if name == "GEOMETRY" {
		return 0, true
	}
	typeNum, ok := wktTags[name]
	return typeNum, ok
}

// formatTypmod returns the text form of a type modifier, such as "(PointZ,4326)", or an empty string when the type
// modifier requires nothing.
func formatTypmod(t typmod) string {
	srid, typeNum := t.srid(), t.typeNum()
	if t < 0 || (srid == 0 && typeNum == 0 && !t.hasZ() && !t.hasM()) {
		return ""
	}
	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(t.typeName())
	if srid != 0 {
		sb.WriteByte(',')
		sb.WriteString(strconv.Itoa(int(srid)))
	}
	sb.WriteByte(')')
	return sb.String()
}

// enforceTypmod checks the geometry against the type modifier. A geometry without an SRID takes the modifier's, a
// single geometry is promoted to the multi-geometry that the modifier requires, and an empty multipoint becomes an
// empty point when the modifier requires a point.
func enforceTypmod(g Geometry, t typmod) (Geometry, error) {
	if t < 0 {
		return g, nil
	}
	modSRID, modType := t.srid(), t.typeNum()
	if modType == pointType && g.Shape.typeNum == multiPointType && g.Shape.isEmpty() {
		g.Shape = shape{typeNum: pointType, ctype: g.Shape.ctype}
	}
	if modSRID > 0 && g.SRID == sridUnknown {
		g.SRID = modSRID
	}
	if modSRID > 0 && modSRID != g.SRID {
		return Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "Geometry SRID (%d) does not match column SRID (%d)", g.SRID, modSRID)
	}
	geomType := g.Shape.typeNum
	if modType > 0 && modType == multiTypeNumber(geomType) {
		g.Shape = asMulti(g.Shape)
		geomType = modType
	}
	if modType > 0 && modType != geomType && !(modType == collectionType && isCollectionType(geomType)) {
		return Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "Geometry type (%s) does not match column type (%s)",
			typeNameForNumber(geomType), typeNameForNumber(modType))
	}
	if t.hasZ() && !g.Shape.hasZ() {
		return Geometry{}, pgerror.New(pgcode.InvalidParameterValue, "Column has Z dimension but geometry does not")
	}
	if g.Shape.hasZ() && !t.hasZ() {
		return Geometry{}, pgerror.New(pgcode.InvalidParameterValue, "Geometry has Z dimension but column does not")
	}
	if t.hasM() && !g.Shape.hasM() {
		return Geometry{}, pgerror.New(pgcode.InvalidParameterValue, "Column has M dimension but geometry does not")
	}
	if g.Shape.hasM() && !t.hasM() {
		return Geometry{}, pgerror.New(pgcode.InvalidParameterValue, "Geometry has M dimension but column does not")
	}
	return g, nil
}

// multiTypeNumber returns the multi-geometry type number that holds the given single type, or the type itself when
// it is not a single type.
func multiTypeNumber(typeNum int) int {
	switch typeNum {
	case pointType:
		return multiPointType
	case lineType:
		return multiLineType
	case polygonType:
		return multiPolygonType
	case circStringType, compoundType:
		return multiCurveType
	case curvePolyType:
		return multiSurfaceType
	case triangleType:
		return tinType
	default:
		return typeNum
	}
}

// isCollectionType returns whether the type number is a multi-geometry or a geometry collection.
func isCollectionType(typeNum int) bool {
	return typeNum == multiPointType || typeNum == multiLineType || typeNum == multiPolygonType || typeNum == collectionType
}

// asMulti wraps a single shape in the matching multi-geometry, and returns any other shape unchanged. An empty single
// shape becomes an empty multi-geometry.
func asMulti(s shape) shape {
	multi := multiTypeNumber(s.typeNum)
	if multi == s.typeNum {
		return s
	}
	if s.isEmpty() {
		return shape{typeNum: multi, ctype: s.ctype}
	}
	return shape{typeNum: multi, ctype: s.ctype, members: []shape{s}}
}

// clampSRID replaces an SRID that PostGIS cannot store, sending the same notice that PostGIS sends. Negative SRIDs
// become zero, and SRIDs above the maximum are folded into the range reserved for them.
func clampSRID(ctx *sql.Context, srid int32) int32 {
	if srid < 0 {
		sendNotice(ctx, "NOTICE", fmt.Sprintf("SRID value %d converted to the officially unknown SRID value %d", srid, sridUnknown))
		return sridUnknown
	}
	if srid > sridMaximum {
		clamped := sridUserMaximum + 1 + srid%(sridMaximum-sridUserMaximum-1)
		sendNotice(ctx, "NOTICE", fmt.Sprintf("SRID value %d > SRID_MAXIMUM converted to %d", srid, clamped))
		return clamped
	}
	return srid
}

// checkLatLongSRID returns an error unless the SRID is unknown or names a longitude/latitude coordinate system.
func checkLatLongSRID(srid int32) error {
	if srid == sridUnknown || srid == sridDefault {
		return nil
	}
	if found, ok := databaseSRSFor(srid); ok && found.geographic {
		return nil
	}
	return pgerror.New(pgcode.InvalidParameterValue, "Only lon/lat coordinate systems are supported in geography.")
}

// sendNotice sends a notice of the given severity, either "NOTICE" or "WARNING", to the client.
func sendNotice(ctx *sql.Context, severity string, message string) {
	if ctx == nil {
		return
	}
	code := pgcode.SuccessfulCompletion
	if severity == "WARNING" {
		code = pgcode.Warning
	}
	dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{
		Severity: severity,
		Code:     code.String(),
		Message:  message,
	})
}
