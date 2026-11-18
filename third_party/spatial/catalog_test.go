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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The expected values come from PROJ 9.8.1, whose database the catalog holds.

func TestCatalogContents(t *testing.T) {
	counts := map[string]int{}
	for _, entry := range catalog().entries {
		counts[entry.Authority]++
	}
	require.Equal(t, map[string]int{"EPSG": 7741, "ESRI": 2991, "IAU_2015": 2201, "IGNF": 864, "NKG": 2, "OGC": 4, "PROJ": 4}, counts)

	wkt, projString, err := CRSDefinitions("EPSG", "4326")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(wkt, `GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,`))
	require.Equal(t, "+proj=longlat +datum=WGS84 +no_defs +type=crs", projString)
	_, _, err = CRSDefinitions("IAU", "49900")
	require.NoError(t, err)
	_, _, err = CRSDefinitions("EPSG", "7030")
	require.Error(t, err)

	infos, err := DatabaseCRSInfo("EPSG", nil)
	require.NoError(t, err)
	require.Len(t, infos, 7741)
	require.Equal(t, CRSInfo{Authority: "EPSG", Code: "3857", Name: "WGS 84 / Pseudo-Mercator", Kind: ProjectedCRS, AreaValid: true,
		West: -180, South: -85.06, East: 180, North: 85.06}, findInfo(infos, "3857"))
	infos, err = DatabaseCRSInfo("unknown", nil)
	require.NoError(t, err)
	require.Empty(t, infos)
}

func TestCatalogSearch(t *testing.T) {
	codes := func(authority string, west, south, east, north float64) string {
		infos, err := DatabaseCRSInfo(authority, &[4]float64{west, south, east, north})
		require.NoError(t, err)
		var codes []string
		for _, info := range infos {
			codes = append(codes, info.Code)
		}
		return strings.Join(codes, " ")
	}
	require.Equal(t, "3395 3409 3410 3832 3857 4087 6932 6933 8857 8858 8859 10601 27706 32700", codes("EPSG", 170, -20, -170, -10))
	require.Equal(t, "4087 8857", codes("EPSG", -180, -90, 180, 90))
	require.Equal(t, "39910 39911 39915 39916 39920 39921 39925 39926 39930 39931 39935 39936 39940 39941 39945 39946 39950 "+
		"39951 39955 39956 39960 39961 39965 39966 39970 39971 39975 39976 39980 39981 39985 39986 39990 39991", codes("IAU", 0, 0, 1, 1))
	_, err := DatabaseCRSInfo("EPSG", &[4]float64{0, 10, 1, 0})
	require.Error(t, err)
}

func TestReadDefinition(t *testing.T) {
	for _, test := range []struct {
		definition string
		expected   definitionKind
	}{
		{"EPSG:4326", geographic},
		{"epsg:4326", geographic},
		{"EPSG:3857", readable},
		{"EPSG:102100", readable},
		{"IAU:49900", geographic},
		{"EPSG:4326+3855", readable},
		{"EPSG:7030", unreadable},
		{"EPSG:99999999", unreadable},
		{"urn:ogc:def:crs:EPSG::4326", geographic},
		{"urn:ogc:def:crs:OGC:1.3:CRS84", geographic},
		{"http://www.opengis.net/def/crs/EPSG/0/4326", geographic},
		{"OGC:CRS84", geographic},
		{"CRS84", unreadable},
		{"garbage", unreadable},
		{"", unreadable},
		{"+proj=longlat +datum=WGS84 +no_defs", readable},
		{"+proj=longlat +datum=WGS84 +type=crs", geographic},
		{"+proj=longlat +ellps=WGS84 +towgs84=0,0,0 +no_defs +type=crs", readable},
		{"+proj=merc +datum=WGS84", readable},
		{"  +proj=longlat +datum=WGS84", readable},
		{"proj=longlat +datum=WGS84", readable},
		{"+init=epsg:4326", unreadable},
		{"+vunits=m +no_defs +type=crs", unreadable},
		{`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563]],PRIMEM["Greenwich",0],UNIT["degree",0.0174532925199433]]`, geographic},
		{`  geogcs["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563]],PRIMEM["Greenwich",0],UNIT["degree",0.0174532925199433]]`, geographic},
		{`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563],TOWGS84[0,0,0,0,0,0,0]],PRIMEM["Greenwich",0],UNIT["degree",0.0174532925199433]]`, readable},
		{`GEOGCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563]],CS[ellipsoidal,2],AXIS["lat",north],AXIS["lon",east],UNIT["degree",0.0174532925199433]]`, geographic},
		{`GEODCRS["WGS 84",DATUM["World Geodetic System 1984",ELLIPSOID["WGS 84",6378137,298.257223563]],CS[Cartesian,3],AXIS["X",geocentricX],AXIS["Y",geocentricY],AXIS["Z",geocentricZ],UNIT["metre",1]]`, readable},
		{`GEOGCS["broken"`, unreadable},
	} {
		require.Equal(t, test.expected, readDefinition(test.definition), test.definition)
		require.Equal(t, test.expected != unreadable, Reads(test.definition), test.definition)
		require.Equal(t, test.expected == geographic, IsGeographic(test.definition), test.definition)
	}
}

func TestBoundingBox(t *testing.T) {
	box, err := newBoundingBox(1, 2, 1, 2)
	require.NoError(t, err)
	require.Less(t, box.west, 1.0)
	require.Greater(t, box.east, 1.0)
	require.Less(t, box.south, 2.0)
	require.Greater(t, box.north, 2.0)
	crossing := boundingBox{west: 170, south: -10, east: -170, north: 10}
	require.True(t, crossing.contains(boundingBox{west: 175, south: 0, east: 178, north: 1}))
	require.True(t, crossing.contains(boundingBox{west: -175, south: 0, east: -172, north: 1}))
	require.False(t, crossing.contains(boundingBox{west: 0, south: 0, east: 1, north: 1}))
	require.False(t, crossing.contains(boundingBox{west: -180, south: 0, east: 180, north: 1}))
	require.True(t, boundingBox{west: -180, south: -90, east: 180, north: 90}.contains(crossing))
}

func findInfo(infos []CRSInfo, code string) CRSInfo {
	for _, info := range infos {
		if info.Code == code {
			return info
		}
	}
	return CRSInfo{}
}
