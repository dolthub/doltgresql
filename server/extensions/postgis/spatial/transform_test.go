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
	"testing"

	"github.com/stretchr/testify/require"
)

// The expected values come from PROJ 9.8.1, through proj_create_crs_to_crs and proj_normalize_for_visualization.

func TestTransformer(t *testing.T) {
	for _, test := range []struct {
		source, target string
		in, expected   [3]float64
	}{
		{"EPSG:4326", "EPSG:3857", [3]float64{-71.06, 42.36, 0}, [3]float64{-7.910363015770021e+06, 5.215059174610621e+06, 0}},
		{"EPSG:3857", "EPSG:4326", [3]float64{100000, 100000, 0}, [3]float64{0.8983152841195214, 0.8982784828192633, 0}},
		{"EPSG:4326", "EPSG:2249", [3]float64{-71.06, 42.36, 10}, [3]float64{775086.4912744571, 2.9565198880358147e+06, 10}},
		{"EPSG:2249", "EPSG:26986", [3]float64{775086.491274, 2956519.888036, 0}, [3]float64{236246.8349323849, 901149.0641714991, 0}},
		{"EPSG:4326", "EPSG:4978", [3]float64{1, 2, 3}, [3]float64{6.373309762299437e+06, 111246.53570857966, 221104.65000960705}},
		{"EPSG:4326", "ESRI:102003", [3]float64{1, 2, 0}, [3]float64{1.0311643196487408e+07, 1.998603959972244e+06, 0}},
		{"EPSG:4326", "EPSG:32633", [3]float64{15, 42, 10}, [3]float64{500000.00000000134, 4.649776224819178e+06, 10}},
		{"EPSG:32631", "EPSG:3857", [3]float64{426857.988, 5427937.523, 0}, [3]float64{222638.98202608555, 6.274861393302727e+06, 0}},
		{"EPSG:4326", "EPSG:2154", [3]float64{0.2600000000000007, 46.355000000000004, 10}, [3]float64{489353.58775158133, 6.587552204045072e+06, 10}},
		{"EPSG:4326", "EPSG:2046", [3]float64{14.475, -22.915, 10}, [3]float64{53858.5028849053, 2.5352204668124304e+06, 10}},
		{"EPSG:4326", "EPSG:27572", [3]float64{1.6800000000000002, 46.735, 10}, [3]float64{549826.6305660798, 2.192990712181458e+06, 10}},
		{"EPSG:4326", "+proj=merc +datum=WGS84", [3]float64{1, 2, 0}, [3]float64{111319.49079327357, 221194.07716771495, 0}},
		{"+proj=longlat +datum=WGS84", "+proj=merc +datum=WGS84", [3]float64{1, 2, 0}, [3]float64{111319.49079327357, 221194.07716771495, 0}},
		{"EPSG:4326", "+proj=utm +zone=33 +ellps=intl +towgs84=-87,-98,-121 +units=m +no_defs", [3]float64{15, 45, 0}, [3]float64{500072.115096078, 4.9831426533648735e+06, 0}},
		{"EPSG:4326", "+proj=lcc +lat_1=33 +lat_2=45 +lat_0=39 +lon_0=-96 +x_0=0 +y_0=0 +datum=NAD83 +units=us-ft +no_defs", [3]float64{-100, 40, 0}, [3]float64{-1.1143146298961958e+06, 386808.0448114165, 0}},
		{"EPSG:4326", "+proj=longlat +ellps=intl +towgs84=-87,-98,-121 +no_defs", [3]float64{15, 45, 5}, [3]float64{15.000914959623794, 45.00089593879195, 5}},
		{"EPSG:4326", "+proj=aea +lat_1=29.5 +lat_2=45.5 +lat_0=37.5 +lon_0=-96 +x_0=0 +y_0=0 +ellps=GRS80 +datum=NAD83 +units=m +no_defs", [3]float64{-90, 35, 0}, [3]float64{542742.2705864373, -262846.5286918778, 0}},
		{"EPSG:4326", "+proj=tmerc +R=6371000 +lon_0=10", [3]float64{12, 45, 0}, [3]float64{157253.36748939555, 5.00571281071217e+06, 0}},
		{"EPSG:4326", "+proj=lcc +lat_1=45 +R=6371000", [3]float64{12, 40, 0}, [3]float64{987394.36558845, 4.239873867017761e+06, 0}},
		{"EPSG:4326", "+proj=merc +R=6371000 +lat_ts=30", [3]float64{12, 40, 0}, [3]float64{1.1555715749536203e+06, 4.2093142176400265e+06, 0}},
		{"EPSG:4326", "+proj=aea +lat_1=20 +lat_2=60 +R=6371000", [3]float64{12, 40, 0}, [3]float64{959347.1743949528, 4.465633148874216e+06, 0}},
		{"EPSG:4326", "EPSG:32631", [3]float64{200, 49, 0}, [3]float64{-740646.4568871276, 1.4427637237141522e+07, 0}},
	} {
		transformer, err := NewTransformer(test.source, test.target)
		require.NoError(t, err, "%s -> %s", test.source, test.target)
		coords := test.in[:]
		require.NoError(t, transformer.Transform(coords, 3, true))
		for i := range coords {
			require.InDelta(t, test.expected[i], coords[i], 1e-8, "%s -> %s", test.source, test.target)
		}
		transformer.Close()
	}
}

func TestTransformerErrors(t *testing.T) {
	transformer, err := NewTransformer("EPSG:4326", "EPSG:3857")
	require.NoError(t, err)
	coords := []float64{200, 100}
	require.Equal(t, TransformError{Message: "Invalid coordinate", Number: 2049}, transformer.Transform(coords, 2, false))
	transformer, err = NewTransformer("EPSG:4326", "EPSG:32631")
	require.NoError(t, err)
	require.Equal(t, TransformError{Message: "Invalid coordinate", Number: 2049}, transformer.Transform([]float64{2, 95}, 2, false))

	// PROJ only chooses the datum shift of these while transforming, and the stereographic projection is not ported
	for _, target := range []string{"EPSG:27700", "EPSG:4269", "+proj=stere +lat_0=90 +datum=WGS84", "+proj=longlat +datum=NAD27"} {
		_, err = NewTransformer("EPSG:4326", target)
		require.ErrorIs(t, err, ErrUnsupportedTransformation, target)
	}
	_, err = NewTransformer("EPSG:27572", "+proj=longlat +ellps=intl +towgs84=-87,-98,-121")
	require.ErrorIs(t, err, ErrUnsupportedTransformation)
	for _, target := range []string{"garbage", "EPSG:99999999", "+proj=longlat +init=epsg:4326", `GEOGCS["WGS 84"]`} {
		_, err = NewTransformer("EPSG:4326", target)
		require.Error(t, err, target)
		require.NotErrorIs(t, err, ErrUnsupportedTransformation, target)
	}
}

func TestPipelineTransformer(t *testing.T) {
	pipeline := "+proj=pipeline +step +proj=unitconvert +xy_in=deg +xy_out=rad +step +proj=utm +zone=31 +ellps=WGS84"
	transformer, err := NewPipelineTransformer(pipeline, true)
	require.NoError(t, err)
	coords := []float64{2, 49}
	require.NoError(t, transformer.Transform(coords, 2, false))
	require.InDelta(t, 426857.988, coords[0], 1e-3)
	require.InDelta(t, 5427937.523, coords[1], 1e-3)
	transformer, err = NewPipelineTransformer(pipeline, false)
	require.NoError(t, err)
	require.NoError(t, transformer.Transform(coords, 2, false))
	require.InDelta(t, 2, coords[0], 1e-9)
	require.InDelta(t, 49, coords[1], 1e-9)

	// An operation whose input is in radians takes degrees
	transformer, err = NewPipelineTransformer("+proj=utm +zone=31 +ellps=WGS84", true)
	require.NoError(t, err)
	coords = []float64{2, 49}
	require.NoError(t, transformer.Transform(coords, 2, false))
	require.InDelta(t, 426857.988, coords[0], 1e-3)

	// A coordinate reference system is read, but transforming with it fails
	transformer, err = NewPipelineTransformer("EPSG:4326", true)
	require.NoError(t, err)
	require.Error(t, transformer.Transform([]float64{2, 49}, 2, false))
	_, err = NewPipelineTransformer("garbage", true)
	require.Error(t, err)
	_, err = NewPipelineTransformer("+proj=pipeline +step +proj=stere +lat_0=90", true)
	require.ErrorIs(t, err, ErrUnsupportedTransformation)
}

func TestDMS(t *testing.T) {
	radians, ok := dmstor(`2d20'14.025"E`)
	require.True(t, ok)
	require.Equal(t, 2*degToRad+20*.0002908882086657216+14.025*.0000048481368110953599, radians)
	radians, ok = dmstor(`74d04'51.3"W`)
	require.True(t, ok)
	require.Less(t, radians, 0.0)
	radians, ok = dmstor("41.5")
	require.True(t, ok)
	require.Equal(t, 41.5*degToRad, radians)
	_, ok = dmstor("north")
	require.False(t, ok)
}
