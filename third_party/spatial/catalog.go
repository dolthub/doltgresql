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

package spatial

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
)

// projDataVersion is the version of PROJ whose database the catalog was generated from.
const projDataVersion = "9.8.1"

// crsEntry is a coordinate reference system of the catalog.
type crsEntry struct {
	CRSInfo
	celestialBody string
	wkt           string
	projString    string
	pipeline      string
}

// crsCatalog holds the catalog's systems in the order that PROJ lists them, along with an index of them by authority
// and code.
type crsCatalog struct {
	entries []crsEntry
	byCode  map[string]map[string]int
}

// versionedAuthorities maps an authority name to the versioned names that PROJ's database files its systems under.
var versionedAuthorities = map[string][]string{"IAU": {"IAU_2015"}}

// catalog returns the parsed catalog, which is parsed on first use.
var catalog = sync.OnceValue(func() *crsCatalog {
	c := &crsCatalog{byCode: make(map[string]map[string]int)}
	for _, line := range strings.Split(catalogText(), "\n") {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 14 {
			panic(errors.Errorf("malformed coordinate reference system catalog line: %q", line))
		}
		kind, _ := strconv.Atoi(fields[2])
		parse := func(s string) float64 {
			v, _ := strconv.ParseFloat(s, 64)
			return v
		}
		entry := crsEntry{
			CRSInfo: CRSInfo{
				Authority:  fields[0],
				Code:       fields[1],
				Name:       fields[10],
				Kind:       CRSKind(kind),
				Deprecated: fields[3] == "1",
				AreaValid:  fields[4] == "1",
				West:       parse(fields[5]),
				South:      parse(fields[6]),
				East:       parse(fields[7]),
				North:      parse(fields[8]),
			},
			celestialBody: fields[9],
			wkt:           fields[11],
			projString:    fields[12],
			pipeline:      fields[13],
		}
		codes, ok := c.byCode[entry.Authority]
		if !ok {
			codes = make(map[string]int)
			c.byCode[entry.Authority] = codes
		}
		codes[entry.Code] = len(c.entries)
		c.entries = append(c.entries, entry)
	}
	return c
})

// lookUp returns the system filed under the given authority and code.
func (c *crsCatalog) lookUp(authority string, code string) (*crsEntry, bool) {
	if i, ok := c.byCode[authority][code]; ok {
		return &c.entries[i], true
	}
	return nil, false
}

// authorities returns the authorities that PROJ searches for the given authority name: its versioned names, or the
// name itself when it has none.
func authorities(authority string) []string {
	if versioned, ok := versionedAuthorities[authority]; ok {
		return versioned
	}
	return []string{authority}
}

// DatabaseCRSInfo returns the coordinate reference systems of the given authority in PROJ's database. When `within`
// holds a west, south, east, and north bound, it returns only the projected systems of Earth that are not deprecated
// and whose area of use holds those bounds.
func DatabaseCRSInfo(authority string, within *[4]float64) ([]CRSInfo, error) {
	var bounds boundingBox
	if within != nil {
		var err error
		if bounds, err = newBoundingBox(within[0], within[1], within[2], within[3]); err != nil {
			return nil, err
		}
	}
	c := catalog()
	var infos []CRSInfo
	for _, name := range authorities(authority) {
		for i := range c.entries {
			entry := &c.entries[i]
			if name != "" && entry.Authority != name {
				continue
			}
			if within != nil {
				if entry.Kind != ProjectedCRS || entry.Deprecated || !entry.AreaValid || entry.celestialBody != "Earth" {
					continue
				}
				if !entry.areaContains(within, bounds) {
					continue
				}
			}
			infos = append(infos, entry.CRSInfo)
		}
	}
	return infos, nil
}

// areaContains returns whether the system's area of use holds the given bounds, as PROJ decides it: by comparing the
// bounds as given when neither box crosses the antimeridian, and otherwise by comparing the widened boxes.
func (e *crsEntry) areaContains(within *[4]float64, bounds boundingBox) bool {
	if e.West <= e.East && within[0] <= within[2] {
		return within[0] >= e.West && within[2] <= e.East && within[1] >= e.South && within[3] <= e.North
	}
	area, err := newBoundingBox(e.West, e.South, e.East, e.North)
	if err != nil {
		return false
	}
	return area.contains(bounds)
}

// CRSDefinitions returns the WKT, as GDAL writes it, and the PROJ string of a coordinate reference system in PROJ's
// database. The PROJ string is empty when PROJ cannot express the system as one.
func CRSDefinitions(authority string, code string) (wkt string, projString string, err error) {
	for _, name := range authorities(authority) {
		if entry, ok := catalog().lookUp(name, code); ok {
			return entry.wkt, entry.projString, nil
		}
	}
	return "", "", errors.Errorf("crs not found: %s:%s", authority, code)
}

// PROJVersion returns the version of PROJ whose database the catalog holds.
func PROJVersion() string {
	return projDataVersion
}

// PROJSettings returns whether PROJ may download data over the network, the address it would download from, the
// directory it writes downloaded data to, and the location of its database. Nothing is ever downloaded, and the
// database is built in.
func PROJSettings() (networkEnabled bool, urlEndpoint string, writableDirectory string, databasePath string) {
	return false, "https://cdn.proj.org", "", ""
}

// definitionKind is how PROJ reads a definition.
type definitionKind uint8

const (
	// unreadable definitions are those that PROJ cannot read.
	unreadable definitionKind = iota
	// readable definitions are those that PROJ reads as something other than a geographic system.
	readable
	// geographic definitions are those that PROJ reads as a geographic system.
	geographic
)

// Reads returns whether PROJ can read the given definition, which may be an authority and code, a URN or URL naming
// one, WKT, or a PROJ string. Definitions that PROJ only reads by looking up an object's name, such as "WGS84", are
// not read.
func Reads(definition string) bool {
	return readDefinition(definition) != unreadable
}

// IsGeographic returns whether PROJ reads the given definition as a geographic coordinate reference system.
func IsGeographic(definition string) bool {
	return readDefinition(definition) == geographic
}

// readDefinition returns how PROJ reads the given definition.
func readDefinition(definition string) definitionKind {
	definition = strings.TrimLeft(definition, " \t\n\r")
	switch {
	case definition == "":
		return unreadable
	case strings.HasPrefix(definition, "+") || strings.HasPrefix(definition, "proj="):
		return readProjString(definition)
	case strings.Contains(definition, "["):
		return readWKT(definition)
	}
	if authority, code, ok := urnAuthorityCode(definition); ok {
		return readAuthorityCode(authority, code)
	}
	if strings.Contains(definition, " ") {
		return unreadable
	}
	if parts := strings.Split(definition, ":"); len(parts) == 2 {
		return readAuthorityCode(parts[0], parts[1])
	}
	return unreadable
}

// readAuthorityCode returns how PROJ reads a system named by an authority and code, matching the authority the way
// PROJ does: exactly, then through its versioned names, and then ignoring case. A code joining two codes with '+'
// names a compound system.
func readAuthorityCode(authority string, code string) definitionKind {
	c := catalog()
	kindOf := func(entry *crsEntry) definitionKind {
		if entry.Kind == Geographic2DCRS || entry.Kind == Geographic3DCRS {
			return geographic
		}
		return readable
	}
	if entry, ok := c.lookUp(authority, code); ok {
		return kindOf(entry)
	}
	if strings.EqualFold(authority, "EPSG") && code == "102100" {
		if entry, ok := c.lookUp("ESRI", code); ok {
			return kindOf(entry)
		}
	}
	for _, name := range versionedAuthorities[authority] {
		if entry, ok := c.lookUp(name, code); ok {
			return kindOf(entry)
		}
	}
	for name := range c.byCode {
		if !strings.EqualFold(name, authority) {
			continue
		}
		if entry, ok := c.lookUp(name, code); ok {
			return kindOf(entry)
		}
		if first, second, ok := strings.Cut(code, "+"); ok && !strings.Contains(second, "+") {
			_, firstOk := c.lookUp(name, first)
			_, secondOk := c.lookUp(name, second)
			if firstOk && secondOk {
				return readable
			}
		}
	}
	return unreadable
}

// urnAuthorityCode returns the authority and code that an OGC URN or URL names, such as "urn:ogc:def:crs:EPSG::4326"
// or "http://www.opengis.net/def/crs/EPSG/0/4326".
func urnAuthorityCode(definition string) (authority string, code string, ok bool) {
	lower := strings.ToLower(definition)
	for _, prefix := range []string{"urn:ogc:def:crs:", "urn:x-ogc:def:crs:"} {
		if strings.HasPrefix(lower, prefix) {
			parts := strings.Split(definition[len(prefix):], ":")
			if len(parts) == 3 {
				return parts[0], parts[2], true
			}
			return "", "", false
		}
	}
	for _, prefix := range []string{"http://www.opengis.net/def/crs/", "https://www.opengis.net/def/crs/"} {
		if strings.HasPrefix(lower, prefix) {
			parts := strings.Split(definition[len(prefix):], "/")
			if len(parts) == 3 {
				return parts[0], parts[2], true
			}
			return "", "", false
		}
	}
	return "", "", false
}

// readWKT returns how PROJ reads a WKT definition, from the keyword of its outermost node. A WKT1 geographic system
// that holds TOWGS84 is read as a bound system, and a WKT2 geodetic system is only geographic with an ellipsoidal
// coordinate system.
func readWKT(definition string) definitionKind {
	open := strings.IndexAny(definition, "[(")
	keyword := strings.ToUpper(strings.TrimSpace(definition[:open]))
	if !balancedWKT(definition) {
		return unreadable
	}
	upper := strings.ToUpper(definition)
	switch keyword {
	case "GEOGCS":
		if strings.Contains(upper, "TOWGS84") {
			return readable
		}
		return geographic
	case "GEOGCRS", "GEOGRAPHICCRS":
		return geographic
	case "GEODCRS", "GEODETICCRS":
		if strings.Contains(upper, "CS[ELLIPSOIDAL") || strings.Contains(upper, "CS(ELLIPSOIDAL") {
			return geographic
		}
		return readable
	case "PROJCS", "GEOCCS", "COMPD_CS", "VERT_CS", "LOCAL_CS", "FITTED_CS", "PROJCRS", "PROJECTEDCRS", "COMPOUNDCRS",
		"VERTCRS", "VERTICALCRS", "ENGCRS", "ENGINEERINGCRS", "BOUNDCRS", "DERIVEDPROJCRS", "PARAMETRICCRS", "TIMECRS",
		"BASEGEOGCRS":
		return readable
	}
	return unreadable
}

// balancedWKT returns whether the brackets of a WKT definition are balanced outside of its quoted strings.
func balancedWKT(definition string) bool {
	depth, quoted := 0, false
	for _, r := range definition {
		switch {
		case r == '"':
			quoted = !quoted
		case quoted:
		case r == '[' || r == '(':
			depth++
		case r == ']' || r == ')':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0 && !quoted
}

// readProjString returns how PROJ reads a PROJ string. Only a string with +type=crs is a system, which is geographic
// for a longitude and latitude projection without a datum shift. PROJ also rejects +init, which needs files that are
// not available, and strings without a projection.
func readProjString(definition string) definitionKind {
	params := make(map[string]string)
	for _, token := range strings.Fields(definition) {
		key, value, _ := strings.Cut(strings.TrimPrefix(token, "+"), "=")
		if _, ok := params[key]; !ok {
			params[key] = value
		}
	}
	if _, ok := params["init"]; ok {
		return unreadable
	}
	projection, ok := params["proj"]
	if !ok || projection == "" {
		return unreadable
	}
	if params["type"] != "crs" {
		return readable
	}
	switch projection {
	case "longlat", "latlong", "lonlat", "latlon":
		if _, ok := params["towgs84"]; ok {
			return readable
		}
		if _, ok := params["nadgrids"]; ok {
			return readable
		}
		if _, ok := params["geoidgrids"]; ok {
			return readable
		}
		return geographic
	}
	return readable
}

// boundingBox is a geographic bounding box in degrees, which crosses the antimeridian when west is greater than
// east.
type boundingBox struct {
	west, south, east, north float64
}

// newBoundingBox returns the bounding box with the given bounds, widening a box that is reduced to a line or a
// point by the smallest amount, as PROJ does.
func newBoundingBox(west float64, south float64, east float64, north float64) (boundingBox, error) {
	if math.IsNaN(west) || math.IsNaN(south) || math.IsNaN(east) || math.IsNaN(north) {
		return boundingBox{}, errors.New("GeographicBoundingBox::create() does not accept NaN values")
	}
	if south > north {
		return boundingBox{}, errors.New("GeographicBoundingBox::create() does not accept south > north")
	}
	if west == east {
		if west > -180 {
			west = math.Nextafter(west, math.Inf(-1))
		}
		if east < 180 {
			east = math.Nextafter(east, math.Inf(1))
		}
	}
	if south == north {
		if south > -90 {
			south = math.Nextafter(south, math.Inf(-1))
		}
		if north < 90 {
			north = math.Nextafter(north, math.Inf(1))
		}
	}
	return boundingBox{west: west, south: south, east: east, north: north}, nil
}

// contains returns whether the box holds the other box, either of which may cross the antimeridian.
func (b boundingBox) contains(other boundingBox) bool {
	if !(b.south <= other.south && b.north >= other.north) {
		return false
	}
	if b.west == -180 && b.east == 180 {
		return other.west != other.east
	}
	if other.west == -180 && other.east == 180 {
		return false
	}
	if b.west < b.east {
		return other.west < other.east && b.west <= other.west && b.east >= other.east
	}
	if other.west < other.east {
		return other.west >= b.west || other.east <= b.east
	}
	return b.west <= other.west && b.east >= other.east
}
