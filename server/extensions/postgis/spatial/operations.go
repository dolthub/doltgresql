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

// This file ports the coordinate operations of PROJ 9.8.1 that the coordinate reference systems of its database
// most often use: the operations themselves (src/projections, src/conversions, src/transformations) along with the
// preparation and finalization that PROJ wraps around every operation (src/fwd.cpp, src/inv.cpp) and the setup of
// ellipsoids (src/ell_set.cpp). The arithmetic follows PROJ's closely so that results agree with PROJ's.

import (
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
)

// The error numbers that PROJ reports for coordinates that it cannot transform.
const (
	errInvalidCoordinate       = 2049
	errOutsideProjectionDomain = 2050
)

// errorMessages holds the messages that PROJ gives for its error numbers.
var errorMessages = map[int]string{
	errInvalidCoordinate:       "Invalid coordinate",
	errOutsideProjectionDomain: "Point outside of projection domain",
}

// The angular constants that PROJ uses.
const (
	halfPi       = math.Pi / 2
	fortPi       = math.Pi / 4
	degToRad     = 0.017453292519943296
	arcsecToRad  = degToRad / 3600.0
	epsLatitude  = 1e-12
	auxLatOrder  = 6
	epsilon10    = 1.e-10
	tolerance7   = 1.e-7
	authalicSize = 0.01
)

// coordinate is a coordinate of up to four ordinates, as PROJ's PJ_COORD.
type coordinate [4]float64

// failed returns whether the coordinate holds the error value that PROJ gives failed coordinates.
func (c *coordinate) failed() bool {
	return math.IsInf(c[0], 0) || math.IsInf(c[1], 0) || math.IsInf(c[2], 0)
}

// fail replaces the coordinate with the error value.
func (c *coordinate) fail() {
	*c = coordinate{math.Inf(1), math.Inf(1), math.Inf(1), math.Inf(1)}
}

// ioUnits are the units of an operation's input or output, as PROJ's PJ_IO_UNITS.
type ioUnits uint8

const (
	unitsWhatever ioUnits = iota
	unitsClassic
	unitsProjected
	unitsCartesian
	unitsRadians
	unitsDegrees
)

// operation is a coordinate operation of a pipeline, which transforms coordinates forward or in reverse. Each
// returns the PROJ error number of a failure, or zero.
type operation interface {
	forward(c *coordinate, stack *[4][]float64) int
	inverse(c *coordinate, stack *[4][]float64) int
	units() (left ioUnits, right ioUnits)
}

// parameters holds the parameters of an operation, keyed by name, with an empty value for flags.
type parameters map[string]string

// has returns whether the parameter is set.
func (p parameters) has(name string) bool {
	_, ok := p[name]
	return ok
}

// float returns the parameter as a number, or the default when it is not set.
func (p parameters) float(name string, def float64) (float64, error) {
	value, ok := p[name]
	if !ok {
		return def, nil
	}
	f, err := parseNumber(value)
	if err != nil {
		return 0, errors.Errorf("invalid value for %s: %s", name, value)
	}
	return f, nil
}

// angle returns the parameter as an angle in radians, given in degrees or in degrees, minutes, and seconds, or the
// default when it is not set.
func (p parameters) angle(name string, def float64) (float64, error) {
	value, ok := p[name]
	if !ok {
		return def, nil
	}
	radians, ok := dmstor(value)
	if !ok {
		return 0, errors.Errorf("invalid value for %s: %s", name, value)
	}
	return radians, nil
}

// dmsFactors holds the radians in a degree, a minute, and a second, as PROJ's dmstor uses them.
var dmsFactors = [3]float64{degToRad, .0002908882086657216, .0000048481368110953599}

// dmstor is PROJ's dmstor, which parses an angle in degrees, minutes, and seconds, such as 2d20'14.025"E, or in
// decimal degrees, and returns it in radians. It returns false when the angle cannot be parsed.
func dmstor(text string) (float64, bool) {
	s := text
	sign := byte('+')
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		sign = s[0]
		s = s[1:]
	}
	v := 0.
	n := 0
	for nl := 0; nl < 3; nl = n + 1 {
		if len(s) == 0 || !((s[0] >= '0' && s[0] <= '9') || s[0] == '.') {
			break
		}
		end := 0
		for end < len(s) && ((s[end] >= '0' && s[end] <= '9') || s[end] == '.' ||
			((s[end] == 'e' || s[end] == 'E') && end+1 < len(s) && (s[end+1] == '-' || s[end+1] == '+' || (s[end+1] >= '0' && s[end+1] <= '9'))) ||
			((s[end] == '-' || s[end] == '+') && end > 0 && (s[end-1] == 'e' || s[end-1] == 'E'))) {
			end++
		}
		tv, err := strconv.ParseFloat(s[:end], 64)
		if err != nil {
			return 0, false
		}
		s = s[end:]
		adv := 1
		switch {
		case len(s) > 0 && (s[0] == 'D' || s[0] == 'd' || s[0] == 0xb0):
			n = 0
		case len(s) > 0 && s[0] == '\'':
			n = 1
		case len(s) > 0 && s[0] == '"':
			n = 2
		case strings.HasPrefix(s, "\xc2\xb0"):
			n = 0
			adv = 2
		case len(s) > 0 && (s[0] == 'r' || s[0] == 'R'):
			if nl != 0 {
				return 0, false
			}
			s = s[1:]
			v = tv
			n = 4
			continue
		default:
			v += tv * dmsFactors[nl]
			n = 4
			continue
		}
		if n < nl {
			return 0, false
		}
		v += tv * dmsFactors[n]
		s = s[adv:]
	}
	if len(s) > 0 {
		if i := strings.IndexByte("NnEeSsWw", s[0]); i >= 0 {
			if i >= 4 {
				sign = '-'
			} else {
				sign = '+'
			}
			s = s[1:]
		}
	}
	if s != "" {
		return 0, false
	}
	if sign == '-' {
		v = -v
	}
	return v, true
}

// primeMeridians holds PROJ's named prime meridians.
var primeMeridians = map[string]string{
	"greenwich": "0dE", "lisbon": "9d07'54.862\"W", "paris": "2d20'14.025\"E", "bogota": "74d04'51.3\"W",
	"madrid": "3d41'16.58\"W", "rome": "12d27'8.4\"E", "bern": "7d26'22.5\"E", "jakarta": "106d48'27.79\"E",
	"ferro": "17d40'W", "brussels": "4d22'4.71\"E", "stockholm": "18d3'29.8\"E", "athens": "23d42'58.815\"E",
	"oslo": "10d43'22.5\"E", "copenhagen": "12d34'40.35\"E",
}

// primeMeridian returns the longitude of the parameters' prime meridian from Greenwich in radians.
func primeMeridian(p parameters) (float64, error) {
	name, ok := p["pm"]
	if !ok {
		return 0, nil
	}
	if definition, ok := primeMeridians[name]; ok {
		name = definition
	}
	radians, ok := dmstor(name)
	if !ok {
		return 0, errors.Errorf("invalid value for pm: %s", p["pm"])
	}
	return radians, nil
}

// parseNumber parses a number, which may be a fraction such as 1/3, as PROJ's pj_atof does.
func parseNumber(value string) (float64, error) {
	if numerator, denominator, ok := strings.Cut(value, "/"); ok {
		n, err := strconv.ParseFloat(numerator, 64)
		if err != nil {
			return 0, err
		}
		d, err := strconv.ParseFloat(denominator, 64)
		if err != nil {
			return 0, err
		}
		return n / d, nil
	}
	return strconv.ParseFloat(value, 64)
}

// ellipsoid holds an ellipsoid's parameters, as PROJ derives them in pj_calc_ellipsoid_params.
type ellipsoid struct {
	a, es, e, f, n, b, ra, oneEs, e2s float64
}

// ellipsoids holds PROJ's named ellipsoids, as their semimajor axis and either their reverse flattening or their
// semiminor axis, the latter prefixed with "b=".
var ellipsoids = map[string][2]string{
	"MERIT": {"6378137.0", "298.257"}, "SGS85": {"6378136.0", "298.257"}, "GRS80": {"6378137.0", "298.257222101"},
	"IAU76": {"6378140.0", "298.257"}, "airy": {"6377563.396", "299.3249646"}, "APL4.9": {"6378137.0", "298.25"},
	"NWL9D": {"6378145.0", "298.25"}, "mod_airy": {"6377340.189", "b=6356034.446"}, "andrae": {"6377104.43", "300.0"},
	"danish": {"6377019.2563", "300.0"}, "aust_SA": {"6378160.0", "298.25"}, "GRS67": {"6378160.0", "298.2471674270"},
	"GSK2011": {"6378136.5", "298.2564151"}, "bessel": {"6377397.155", "299.1528128"},
	"bess_nam": {"6377483.865", "299.1528128"}, "clrk66": {"6378206.4", "b=6356583.8"},
	"clrk80": {"6378249.145", "293.4663"}, "clrk80ign": {"6378249.2", "293.4660212936269"},
	"CPM": {"6375738.7", "334.29"}, "delmbr": {"6376428.", "311.5"}, "engelis": {"6378136.05", "298.2566"},
	"evrst30": {"6377276.345", "300.8017"}, "evrst48": {"6377304.063", "300.8017"},
	"evrst56": {"6377301.243", "300.8017"}, "evrst69": {"6377295.664", "300.8017"},
	"evrstSS": {"6377298.556", "300.8017"}, "fschr60": {"6378166.", "298.3"}, "fschr60m": {"6378155.", "298.3"},
	"fschr68": {"6378150.", "298.3"}, "helmert": {"6378200.", "298.3"}, "hough": {"6378270.0", "297."},
	"intl": {"6378388.0", "297."}, "krass": {"6378245.0", "298.3"}, "kaula": {"6378163.", "298.24"},
	"lerch": {"6378139.", "298.257"}, "mprts": {"6397300.", "191."}, "new_intl": {"6378157.5", "b=6356772.2"},
	"plessis": {"6376523.", "b=6355863."}, "PZ90": {"6378136.0", "298.25784"}, "SEasia": {"6378155.0", "b=6356773.3205"},
	"walbeck": {"6376896.0", "b=6355834.8467"}, "WGS60": {"6378165.0", "298.3"}, "WGS66": {"6378145.0", "298.25"},
	"WGS72": {"6378135.0", "298.26"}, "WGS84": {"6378137.0", "298.257223563"}, "sphere": {"6370997.0", "b=6370997.0"},
}

// datum is one of PROJ's named datums: its ellipsoid and its shift to WGS 84, as either towgs84 or nadgrids.
type datum struct {
	ellipsoid string
	towgs84   string
	nadgrids  string
}

// datums holds PROJ's named datums.
var datums = map[string]datum{
	"WGS84":         {ellipsoid: "WGS84", towgs84: "0,0,0"},
	"GGRS87":        {ellipsoid: "GRS80", towgs84: "-199.87,74.79,246.62"},
	"NAD83":         {ellipsoid: "GRS80", towgs84: "0,0,0"},
	"NAD27":         {ellipsoid: "clrk66", nadgrids: "@conus,@alaska,@ntv2_0.gsb,@ntv1_can.dat"},
	"potsdam":       {ellipsoid: "bessel", nadgrids: "@BETA2007.gsb"},
	"carthage":      {ellipsoid: "clrk80ign", towgs84: "-263.0,6.0,431.0"},
	"hermannskogel": {ellipsoid: "bessel", towgs84: "577.326,90.129,463.919,5.137,1.474,5.297,2.4232"},
	"ire65":         {ellipsoid: "mod_airy", towgs84: "482.530,-130.596,564.557,-1.042,-0.214,-0.631,8.15"},
	"nzgd49":        {ellipsoid: "intl", towgs84: "59.47,-5.04,187.44,0.47,-0.1,1.024,-4.5993"},
	"OSGB36":        {ellipsoid: "airy", towgs84: "446.448,-125.157,542.060,0.1502,0.2470,0.8421,-20.4894"},
}

// newEllipsoid returns the ellipsoid that the parameters define, which is GRS 1980 when they define none, as in
// PROJ.
func newEllipsoid(p parameters) (ellipsoid, error) {
	for key := range p {
		if strings.HasPrefix(key, "R_") {
			return ellipsoid{}, errors.Wrapf(ErrUnsupportedTransformation, "the %s parameter is not supported", key)
		}
	}
	var a, es, f, b float64
	if r, ok := p["R"]; ok {
		radius, err := parseNumber(r)
		if err != nil || radius <= 0 {
			return ellipsoid{}, errors.Errorf("invalid value for R: %s", r)
		}
		return deriveEllipsoid(radius, 0, 0, 0)
	}
	name := p["ellps"]
	if datumName, ok := p["datum"]; ok && name == "" {
		d, ok := datums[datumName]
		if !ok {
			return ellipsoid{}, errors.Errorf("unknown datum: %s", datumName)
		}
		name = d.ellipsoid
	}
	shapeKey, shapeValue := "", ""
	if name != "" {
		named, ok := ellipsoids[name]
		if !ok {
			return ellipsoid{}, errors.Errorf("unknown ellipsoid: %s", name)
		}
		a, _ = parseNumber(named[0])
		if strings.HasPrefix(named[1], "b=") {
			shapeKey, shapeValue = "b", named[1][2:]
		} else {
			shapeKey, shapeValue = "rf", named[1]
		}
	} else if !p.has("a") {
		a = 6378137.0
		shapeKey, shapeValue = "rf", "298.257222101"
	}
	if value, ok := p["a"]; ok {
		var err error
		if a, err = parseNumber(value); err != nil || a <= 0 {
			return ellipsoid{}, errors.Errorf("invalid value for a: %s", value)
		}
		shapeKey, shapeValue = "", ""
	}
	for _, key := range []string{"rf", "f", "es", "e", "b"} {
		if value, ok := p[key]; ok {
			shapeKey, shapeValue = key, value
			break
		}
	}
	if shapeKey != "" {
		v, err := parseNumber(shapeValue)
		if err != nil {
			return ellipsoid{}, errors.Errorf("invalid value for %s: %s", shapeKey, shapeValue)
		}
		switch shapeKey {
		case "rf":
			if v <= 0 {
				return ellipsoid{}, errors.New("invalid value for rf")
			}
			f = 1 / v
			es = 2*f - f*f
		case "f":
			f = v
			es = 2*f - f*f
		case "es":
			es = v
		case "e":
			es = v * v
		case "b":
			b = v
			if b != a {
				f = (a - b) / a
				es = 2*f - f*f
			}
		}
	}
	return deriveEllipsoid(a, es, f, b)
}

// deriveEllipsoid returns the ellipsoid with the given semimajor axis and squared eccentricity, along with the
// flattening and semiminor axis when they are known, deriving the rest as pj_calc_ellipsoid_params does.
func deriveEllipsoid(a float64, es float64, f float64, b float64) (ellipsoid, error) {
	if !(es >= 0) || es >= 1 {
		return ellipsoid{}, errors.New("invalid eccentricity")
	}
	e := math.Sqrt(es)
	alpha := math.Asin(e)
	e2 := math.Tan(alpha)
	if f == 0 {
		f = 1 - math.Cos(alpha)
	}
	n := math.Pow(math.Tan(alpha/2), 2)
	if b == 0 {
		b = (1 - f) * a
	}
	return ellipsoid{a: a, es: es, e: e, f: f, n: n, b: b, ra: 1 / a, oneEs: 1 - es, e2s: e2 * e2}, nil
}

// adjlon returns the longitude in the range from -pi to pi, as PROJ's adjlon.
func adjlon(longitude float64) float64 {
	if math.Abs(longitude) < math.Pi+1e-12 {
		return longitude
	}
	longitude += math.Pi
	longitude -= 2 * math.Pi * math.Floor(longitude/(2*math.Pi))
	longitude -= math.Pi
	return longitude
}

// projectionFunctions are the forward and inverse functions of a projection, which work in units of the semimajor
// axis and return a PROJ error number on failure.
type projectionFunctions struct {
	fwd func(lam float64, phi float64) (x float64, y float64, err int)
	inv func(x float64, y float64) (lam float64, phi float64, err int)
}

// projection is an operation from longitude and latitude in radians to plane coordinates, with the preparation and
// finalization that PROJ gives classic projections.
type projection struct {
	ell              ellipsoid
	lam0, phi0, k0   float64
	fromGreenwich    float64
	x0, y0           float64
	toMeter, frMeter float64
	over             bool
	radiansOut       bool
	axis             *axisSwap
	projectionFunctions
}

var _ operation = (*projection)(nil)

// units implements the interface operation.
func (p *projection) units() (ioUnits, ioUnits) {
	if p.radiansOut {
		return unitsRadians, unitsRadians
	}
	return unitsRadians, unitsClassic
}

// prepareAngular checks and adjusts angular input as PROJ's fwd_prepare does.
func prepareAngular(c *coordinate, fromGreenwich float64, lam0 float64, over bool) int {
	if c.failed() {
		c.fail()
		return 0
	}
	t := math.Abs(c[1]) - halfPi
	if t > epsLatitude || c[0] > 10 || c[0] < -10 {
		c.fail()
		return errInvalidCoordinate
	}
	if c[1] > halfPi {
		c[1] = halfPi
	}
	if c[1] < -halfPi {
		c[1] = -halfPi
	}
	if !over {
		c[0] = adjlon(c[0])
	}
	c[0] = (c[0] - fromGreenwich) - lam0
	if !over {
		c[0] = adjlon(c[0])
	}
	return 0
}

// finishAngular adjusts angular output as PROJ's inv_finalize does.
func finishAngular(c *coordinate, fromGreenwich float64, lam0 float64, over bool) int {
	if math.IsInf(c[0], 0) {
		c.fail()
		return errOutsideProjectionDomain
	}
	c[0] = c[0] + fromGreenwich + lam0
	if !over {
		c[0] = adjlon(c[0])
	}
	return 0
}

// forward implements the interface operation.
func (p *projection) forward(c *coordinate, _ *[4][]float64) int {
	if err := prepareAngular(c, p.fromGreenwich, p.lam0, p.over); err != 0 || c.failed() {
		return err
	}
	x, y, err := p.fwd(c[0], c[1])
	if err != 0 || math.IsInf(x, 0) || math.IsInf(y, 0) {
		c.fail()
		if err == 0 {
			err = errOutsideProjectionDomain
		}
		return err
	}
	if p.radiansOut {
		c[0], c[1] = x, y
	} else {
		c[0] = p.frMeter * (x*p.ell.a + p.x0)
		c[1] = p.frMeter * (y*p.ell.a + p.y0)
	}
	if p.axis != nil {
		p.axis.forward(c, nil)
	}
	return 0
}

// inverse implements the interface operation.
func (p *projection) inverse(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return errOutsideProjectionDomain
	}
	if p.axis != nil {
		p.axis.inverse(c, nil)
	}
	x, y := c[0], c[1]
	if !p.radiansOut {
		x = (p.toMeter*x - p.x0) * p.ell.ra
		y = (p.toMeter*y - p.y0) * p.ell.ra
	}
	lam, phi, err := p.inv(x, y)
	if err != 0 {
		c.fail()
		return err
	}
	c[0], c[1] = lam, phi
	return finishAngular(c, p.fromGreenwich, p.lam0, p.over)
}

// newProjection returns the projection or angular operation that the parameters name, or false when it is not one
// that this package implements.
func newProjection(name string, p parameters) (operation, bool, error) {
	setup, ok := projectionSetups[name]
	if !ok {
		return nil, false, nil
	}
	if p.has("geoc") || p.has("approx") || (p.has("algo") && p["algo"] != "poder_engsager") {
		return nil, false, nil
	}
	ell, err := newEllipsoid(p)
	if err != nil {
		return nil, true, err
	}
	proj := &projection{ell: ell, k0: 1, toMeter: 1, frMeter: 1, over: p.has("over")}
	if proj.fromGreenwich, err = primeMeridian(p); err != nil {
		return nil, true, err
	}
	if proj.lam0, err = p.angle("lon_0", 0); err != nil {
		return nil, true, err
	}
	if proj.phi0, err = p.angle("lat_0", 0); err != nil {
		return nil, true, err
	}
	if proj.x0, err = p.float("x_0", 0); err != nil {
		return nil, true, err
	}
	if proj.y0, err = p.float("y_0", 0); err != nil {
		return nil, true, err
	}
	for _, key := range []string{"k_0", "k"} {
		if p.has(key) {
			if proj.k0, err = p.float(key, 1); err != nil {
				return nil, true, err
			}
			break
		}
	}
	if units, ok := p["units"]; ok {
		factor, ok := linearUnits[units]
		if !ok {
			return nil, true, errors.Errorf("unknown units: %s", units)
		}
		proj.toMeter, proj.frMeter = factor, 1/factor
	} else if p.has("to_meter") {
		factor, err := p.float("to_meter", 1)
		if err != nil || factor <= 0 {
			return nil, true, errors.New("invalid value for to_meter")
		}
		proj.toMeter, proj.frMeter = factor, 1/factor
	}
	if axis, ok := p["axis"]; ok && axis != "enu" {
		if proj.axis, err = axisFromLetters(axis); err != nil {
			return nil, true, err
		}
	}
	if err = setup(proj, p); err != nil {
		return nil, true, err
	}
	return proj, true, nil
}

// axisFromLetters returns the axis swap of PROJ's +axis parameter, whose letters give the direction of each output
// axis: east or west, north or south, and up or down.
func axisFromLetters(letters string) (*axisSwap, error) {
	if len(letters) != 3 {
		return nil, errors.Errorf("invalid value for axis: %s", letters)
	}
	directions := map[byte]int{'e': 1, 'w': -1, 'n': 2, 's': -2, 'u': 3, 'd': -3}
	order := make([]string, 3)
	for i := range order {
		direction, ok := directions[letters[i]]
		if !ok {
			return nil, errors.Errorf("invalid value for axis: %s", letters)
		}
		order[i] = strconv.Itoa(direction)
	}
	op, err := newAxisSwap(parameters{"order": strings.Join(order, ",")})
	if err != nil {
		return nil, err
	}
	return op.(*axisSwap), nil
}

// projectionSetups holds the setup of every projection that this package implements, keyed by PROJ's name for it.
var projectionSetups = map[string]func(p *projection, params parameters) error{
	"tmerc":   setupTmerc,
	"etmerc":  setupTmerc,
	"utm":     setupUTM,
	"lcc":     setupLCC,
	"merc":    setupMerc,
	"webmerc": setupWebMerc,
	"aea":     setupAEA,
	"longlat": setupLongLat,
	"latlong": setupLongLat,
	"lonlat":  setupLongLat,
	"latlon":  setupLongLat,
}

// setupLongLat sets up PROJ's longlat operation, which leaves longitudes and latitudes in radians alone.
func setupLongLat(p *projection, _ parameters) error {
	p.radiansOut = true
	p.fwd = func(lam float64, phi float64) (float64, float64, int) { return lam, phi, 0 }
	p.inv = func(x float64, y float64) (float64, float64, int) { return x, y, 0 }
	return nil
}

// tsfn is PROJ's pj_tsfn.
func tsfn(phi float64, sinphi float64, e float64) float64 {
	cosphi := math.Cos(phi)
	factor := (1 - sinphi) / cosphi
	if sinphi > 0 {
		factor = cosphi / (1 + sinphi)
	}
	return math.Exp(e*math.Atanh(e*sinphi)) * factor
}

// msfn is PROJ's pj_msfn.
func msfn(sinphi float64, cosphi float64, es float64) float64 {
	return cosphi / math.Sqrt(1.-es*sinphi*sinphi)
}

// sinhpsi2tanphi is PROJ's pj_sinhpsi2tanphi.
func sinhpsi2tanphi(taup float64, e float64) float64 {
	const numit = 5
	rooteps := math.Sqrt(2.220446049250313e-16)
	tol := rooteps / 10
	tmax := 2 / rooteps
	e2m := 1 - e*e
	stol := tol * math.Max(1.0, math.Abs(taup))
	tau := taup / e2m
	if math.Abs(taup) > 70 {
		tau = taup * math.Exp(e*math.Atanh(e))
	}
	if !(math.Abs(tau) < tmax) {
		return tau
	}
	for i := numit; i > 0; i-- {
		tau1 := math.Sqrt(1 + tau*tau)
		sig := math.Sinh(e * math.Atanh(e*tau/tau1))
		taupa := math.Sqrt(1+sig*sig)*tau - sig*tau1
		dtau := (taup - taupa) * (1 + e2m*(tau*tau)) / (e2m * tau1 * math.Sqrt(1+taupa*taupa))
		tau += dtau
		if !(math.Abs(dtau) >= stol) {
			break
		}
	}
	return tau
}

// pjPhi2 is PROJ's pj_phi2.
func pjPhi2(ts0 float64, e float64) float64 {
	return math.Atan(sinhpsi2tanphi((1/ts0-ts0)/2, e))
}

// The auxiliary latitudes that PROJ converts between.
const (
	auxGeographic = 0
	auxRectifying = 3
	auxConformal  = 4
	auxAuthalic   = 5
)

// auxLatCoefficients holds the coefficients of PROJ's pj_auxlat_coeffs, keyed by the input and output auxiliary
// latitudes.
var auxLatCoefficients = map[[2]int][]float64{
	{auxRectifying, auxGeographic}: {3.0 / 2.0, -27.0 / 32.0, 269.0 / 512.0, 21.0 / 16.0, -55.0 / 32.0,
		6759.0 / 4096.0, 151.0 / 96.0, -417.0 / 128.0, 1097.0 / 512.0, -15543.0 / 2560.0, 8011.0 / 2560.0,
		293393.0 / 61440.0},
	{auxConformal, auxGeographic}: {2.0, -2.0 / 3.0, -2.0, 116.0 / 45.0, 26.0 / 45.0, -2854.0 / 675.0, 7.0 / 3.0,
		-8.0 / 5.0, -227.0 / 45.0, 2704.0 / 315.0, 2323.0 / 945.0, 56.0 / 15.0, -136.0 / 35.0, -1262.0 / 105.0,
		73814.0 / 2835.0, 4279.0 / 630.0, -332.0 / 35.0, -399572.0 / 14175.0, 4174.0 / 315.0, -144838.0 / 6237.0,
		601676.0 / 22275.0},
	{auxAuthalic, auxGeographic}: {4.0 / 3.0, 4.0 / 45.0, -16.0 / 35.0, -2582.0 / 14175.0, 60136.0 / 467775.0,
		28112932.0 / 212837625.0, 46.0 / 45.0, 152.0 / 945.0, -11966.0 / 14175.0, -21016.0 / 51975.0,
		251310128.0 / 638512875.0, 3044.0 / 2835.0, 3802.0 / 14175.0, -94388.0 / 66825.0, -8797648.0 / 10945935.0,
		6059.0 / 4725.0, 41072.0 / 93555.0, -1472637812.0 / 638512875.0, 768272.0 / 467775.0,
		455935736.0 / 638512875.0, 4210684958.0 / 1915538625.0},
	{auxGeographic, auxRectifying}: {-3.0 / 2.0, 9.0 / 16.0, -3.0 / 32.0, 15.0 / 16.0, -15.0 / 32.0, 135.0 / 2048.0,
		-35.0 / 48.0, 105.0 / 256.0, 315.0 / 512.0, -189.0 / 512.0, -693.0 / 1280.0, 1001.0 / 2048.0},
	{auxConformal, auxRectifying}: {1.0 / 2.0, -2.0 / 3.0, 5.0 / 16.0, 41.0 / 180.0, -127.0 / 288.0, 7891.0 / 37800.0,
		13.0 / 48.0, -3.0 / 5.0, 557.0 / 1440.0, 281.0 / 630.0, -1983433.0 / 1935360.0, 61.0 / 240.0, -103.0 / 140.0,
		15061.0 / 26880.0, 167603.0 / 181440.0, 49561.0 / 161280.0, -179.0 / 168.0, 6601661.0 / 7257600.0,
		34729.0 / 80640.0, -3418889.0 / 1995840.0, 212378941.0 / 319334400.0},
	{auxGeographic, auxConformal}: {-2.0, 2.0 / 3.0, 4.0 / 3.0, -82.0 / 45.0, 32.0 / 45.0, 4642.0 / 4725.0, 5.0 / 3.0,
		-16.0 / 15.0, -13.0 / 9.0, 904.0 / 315.0, -1522.0 / 945.0, -26.0 / 15.0, 34.0 / 21.0, 8.0 / 5.0,
		-12686.0 / 2835.0, 1237.0 / 630.0, -12.0 / 5.0, -24832.0 / 14175.0, -734.0 / 315.0, 109598.0 / 31185.0,
		444337.0 / 155925.0},
	{auxRectifying, auxConformal}: {-1.0 / 2.0, 2.0 / 3.0, -37.0 / 96.0, 1.0 / 360.0, 81.0 / 512.0,
		-96199.0 / 604800.0, -1.0 / 48.0, -1.0 / 15.0, 437.0 / 1440.0, -46.0 / 105.0, 1118711.0 / 3870720.0,
		-17.0 / 480.0, 37.0 / 840.0, 209.0 / 4480.0, -5569.0 / 90720.0, -4397.0 / 161280.0, 11.0 / 504.0,
		830251.0 / 7257600.0, -4583.0 / 161280.0, 108847.0 / 3991680.0, -20648693.0 / 638668800.0},
	{auxGeographic, auxAuthalic}: {-4.0 / 3.0, -4.0 / 45.0, 88.0 / 315.0, 538.0 / 4725.0, 20824.0 / 467775.0,
		-44732.0 / 2837835.0, 34.0 / 45.0, 8.0 / 105.0, -2482.0 / 14175.0, -37192.0 / 467775.0,
		-12467764.0 / 212837625.0, -1532.0 / 2835.0, -898.0 / 14175.0, 54968.0 / 467775.0,
		100320856.0 / 1915538625.0, 6007.0 / 14175.0, 24496.0 / 467775.0, -5884124.0 / 70945875.0,
		-23356.0 / 66825.0, -839792.0 / 19348875.0, 570284222.0 / 1915538625.0},
}

// polyval is PROJ's pj_polyval.
func polyval(x float64, p []float64, n int) float64 {
	if n < 0 {
		return 0
	}
	y := p[n]
	for n > 0 {
		n--
		y = y*x + p[n]
	}
	return y
}

// auxLatCoeffs is PROJ's pj_auxlat_coeffs, which returns the coefficients to convert the auxiliary latitude `in` to
// `out`.
func auxLatCoeffs(n float64, in int, out int) []float64 {
	coeffs := auxLatCoefficients[[2]int{in, out}]
	f := make([]float64, auxLatOrder)
	o := 0
	d, n2 := n, n*n
	if in <= auxRectifying && out <= auxRectifying {
		for l := 0; l < auxLatOrder; l++ {
			m := (auxLatOrder - l - 1) / 2
			f[l] = d * polyval(n2, coeffs[o:], m)
			o += m + 1
			d *= n
		}
	} else {
		for l := 0; l < auxLatOrder; l++ {
			m := auxLatOrder - l - 1
			f[l] = d * polyval(n, coeffs[o:], m)
			o += m + 1
			d *= n
		}
	}
	return f
}

// clenshaw is PROJ's pj_clenshaw.
func clenshaw(szeta float64, czeta float64, f []float64) float64 {
	u0, u1 := 0.0, 0.0
	x := 2 * (czeta - szeta) * (czeta + szeta)
	for k := len(f); k > 0; {
		k--
		t := x*u0 - u1 + f[k]
		u1 = u0
		u0 = t
	}
	return 2 * szeta * czeta * u0
}

// auxLatConvert is PROJ's pj_auxlat_convert.
func auxLatConvert(zeta float64, f []float64) float64 {
	return zeta + clenshaw(math.Sin(zeta), math.Cos(zeta), f)
}

// auxLatConvertSinCos is PROJ's pj_auxlat_convert with the sine and cosine of the latitude given.
func auxLatConvertSinCos(zeta float64, szeta float64, czeta float64, f []float64) float64 {
	return zeta + clenshaw(szeta, czeta, f)
}

// rectifyingRadius is PROJ's pj_rectifying_radius.
func rectifyingRadius(n float64) float64 {
	return polyval(n*n, []float64{1, 1.0 / 4, 1.0 / 64, 1.0 / 256}, 3) / (1 + n)
}

// clenS is the complex Clenshaw summation of PROJ's transverse Mercator.
func clenS(a []float64, sinArgR float64, cosArgR float64, sinhArgI float64, coshArgI float64) (float64, float64) {
	r := 2 * cosArgR * coshArgI
	i := -2 * sinArgR * sinhArgI
	p := len(a)
	var hi, hi1, hi2, hr1, hr2 float64
	p--
	hr := a[p]
	for p > 0 {
		hr2 = hr1
		hi2 = hi1
		hr1 = hr
		hi1 = hi
		p--
		hr = -hr2 + r*hr1 - i*hi1 + a[p]
		hi = -hi2 + i*hr1 + r*hi1
	}
	r = sinArgR * coshArgI
	i = cosArgR * sinhArgI
	return r*hr - i*hi, r*hi + i*hr
}

// setupTmerc sets up PROJ's transverse Mercator, which uses the Poder/Engsager algorithm on ellipsoids and the
// Evenden/Snyder algorithm on spheres.
func setupTmerc(p *projection, _ parameters) error {
	if p.ell.es == 0 {
		setupSphericalTmerc(p)
		return nil
	}
	n := p.ell.n
	cgb := auxLatCoeffs(n, auxConformal, auxGeographic)
	cbg := auxLatCoeffs(n, auxGeographic, auxConformal)
	qn := p.k0 * rectifyingRadius(n)
	utg := auxLatCoeffs(n, auxRectifying, auxConformal)
	gtu := auxLatCoeffs(n, auxConformal, auxRectifying)
	z := auxLatConvert(p.phi0, cbg)
	zb := -qn * auxLatConvert(z, gtu)
	p.fwd = func(lam float64, phi float64) (float64, float64, int) {
		cn := auxLatConvert(phi, cbg)
		sinCn, cosCn := math.Sin(cn), math.Cos(cn)
		sinCe, cosCe := math.Sin(lam), math.Cos(lam)
		cosCnCosCe := cosCn * cosCe
		cn = math.Atan2(sinCn, cosCnCosCe)
		invDenomTanCe := 1. / math.Hypot(sinCn, cosCnCosCe)
		tanCe := sinCe * cosCn * invDenomTanCe
		ce := math.Asinh(tanCe)
		twoInvDenomTanCe := 2 * invDenomTanCe
		twoInvDenomTanCeSquare := twoInvDenomTanCe * invDenomTanCe
		tmpR := cosCnCosCe * twoInvDenomTanCeSquare
		sinArgR := sinCn * tmpR
		cosArgR := cosCnCosCe*tmpR - 1
		sinhArgI := tanCe * twoInvDenomTanCe
		coshArgI := twoInvDenomTanCeSquare - 1
		dCn, dCe := clenS(gtu, sinArgR, cosArgR, sinhArgI, coshArgI)
		cn += dCn
		ce += dCe
		if math.Abs(ce) <= 2.623395162778 {
			return qn * ce, qn*cn + zb, 0
		}
		return math.Inf(1), math.Inf(1), errOutsideProjectionDomain
	}
	p.inv = func(x float64, y float64) (float64, float64, int) {
		cn := (y - zb) / qn
		ce := x / qn
		if !(math.Abs(ce) <= 2.623395162778) {
			return math.Inf(1), math.Inf(1), errOutsideProjectionDomain
		}
		sinArgR, cosArgR := math.Sin(2*cn), math.Cos(2*cn)
		exp2Ce := math.Exp(2 * ce)
		halfInvExp2Ce := 0.5 / exp2Ce
		sinhArgI := 0.5*exp2Ce - halfInvExp2Ce
		coshArgI := 0.5*exp2Ce + halfInvExp2Ce
		dCn, dCe := clenS(utg, sinArgR, cosArgR, sinhArgI, coshArgI)
		cn += dCn
		ce += dCe
		sinCn, cosCn := math.Sin(cn), math.Cos(cn)
		sinhCe := math.Sinh(ce)
		ce = math.Atan2(sinhCe, cosCn)
		modulusCe := math.Hypot(sinhCe, cosCn)
		rr := math.Hypot(sinCn, modulusCe)
		cn = math.Atan2(sinCn, modulusCe)
		return ce, auxLatConvertSinCos(cn, sinCn/rr, modulusCe/rr, cgb), 0
	}
	return nil
}

// setupSphericalTmerc sets up PROJ's spherical transverse Mercator.
func setupSphericalTmerc(p *projection) {
	esp := p.k0
	ml0 := .5 * esp
	p.fwd = func(lam float64, phi float64) (float64, float64, int) {
		cosphi := math.Cos(phi)
		b := cosphi * math.Sin(lam)
		if math.Abs(math.Abs(b)-1.) <= epsilon10 {
			return 0, 0, errOutsideProjectionDomain
		}
		x := ml0 * math.Log((1.+b)/(1.-b))
		var y float64
		if cosphi == 1.0 {
			if lam < -halfPi || lam > halfPi {
				y = math.Pi
			}
		} else {
			y = cosphi * math.Cos(lam) / math.Sqrt(1.-b*b)
			b = math.Abs(y)
			if b >= 1. {
				if (b - 1.) > epsilon10 {
					return 0, 0, errOutsideProjectionDomain
				}
				y = 0.
			} else {
				y = math.Acos(y)
			}
		}
		if phi < 0. {
			y = -y
		}
		return x, esp * (y - p.phi0), 0
	}
	p.inv = func(x float64, y float64) (float64, float64, int) {
		h := math.Exp(x / esp)
		if h == 0 {
			return math.Inf(1), math.Inf(1), errOutsideProjectionDomain
		}
		g := .5 * (h - 1./h)
		d := p.phi0 + y/esp
		h = math.Cos(d)
		phi := math.Copysign(math.Asin(math.Sqrt((1.-h*h)/(1.+g*g))), d)
		lam := 0.
		if g != 0.0 || h != 0.0 {
			lam = math.Atan2(g, h)
		}
		return lam, phi, 0
	}
}

// setupUTM sets up PROJ's UTM, which is transverse Mercator with the zone's parameters.
func setupUTM(p *projection, params parameters) error {
	if p.ell.es == 0 {
		return errors.New("invalid value for eccentricity: it should not be zero")
	}
	p.y0 = 0
	if params.has("south") {
		p.y0 = 10000000.
	}
	p.x0 = 500000.
	var zone int
	if value, ok := params["zone"]; ok {
		z, err := strconv.Atoi(value)
		if err != nil || z <= 0 || z > 60 {
			return errors.New("invalid value for zone")
		}
		zone = z - 1
	} else {
		zone = int(math.Round(math.Floor((adjlon(p.lam0) + math.Pi) * 30. / math.Pi)))
		zone = max(0, min(zone, 59))
	}
	p.lam0 = (float64(zone)+.5)*math.Pi/30. - math.Pi
	p.k0 = 0.9996
	p.phi0 = 0.
	return setupTmerc(p, params)
}

// setupLCC sets up PROJ's Lambert Conformal Conic.
func setupLCC(p *projection, params parameters) error {
	phi1, err := params.angle("lat_1", 0)
	if err != nil {
		return err
	}
	phi2 := phi1
	if params.has("lat_2") {
		if phi2, err = params.angle("lat_2", 0); err != nil {
			return err
		}
	} else if !params.has("lat_0") {
		p.phi0 = phi1
	}
	if math.Abs(phi1+phi2) < epsilon10 {
		return errors.New("invalid value for lat_1 and lat_2: |lat_1 + lat_2| should be > 0")
	}
	sinphi := math.Sin(phi1)
	n := sinphi
	cosphi := math.Cos(phi1)
	if math.Abs(cosphi) < epsilon10 || math.Abs(phi1) >= halfPi || math.Abs(math.Cos(phi2)) < epsilon10 || math.Abs(phi2) >= halfPi {
		return errors.New("invalid value for lat_1 or lat_2: it should be within 90 degrees")
	}
	secant := math.Abs(phi1-phi2) >= epsilon10
	ell := p.ell
	var c, rho0 float64
	if ell.es != 0. {
		m1 := msfn(sinphi, cosphi, ell.es)
		ml1 := tsfn(phi1, sinphi, ell.e)
		if secant {
			sinphi = math.Sin(phi2)
			n = math.Log(m1 / msfn(sinphi, math.Cos(phi2), ell.es))
			if n == 0 {
				return errors.New("invalid value for eccentricity")
			}
			ml2 := tsfn(phi2, sinphi, ell.e)
			denom := math.Log(ml1 / ml2)
			if denom == 0 {
				return errors.New("invalid value for eccentricity")
			}
			n /= denom
		}
		rho0 = m1 * math.Pow(ml1, -n) / n
		c = rho0
		if math.Abs(math.Abs(p.phi0)-halfPi) < epsilon10 {
			rho0 *= 0.
		} else {
			rho0 *= math.Pow(tsfn(p.phi0, math.Sin(p.phi0), ell.e), n)
		}
	} else {
		if secant {
			n = math.Log(cosphi/math.Cos(phi2)) / math.Log(math.Tan(fortPi+.5*phi2)/math.Tan(fortPi+.5*phi1))
		}
		if n == 0 {
			return errors.New("invalid value for lat_1 and lat_2: |lat_1 + lat_2| should be > 0")
		}
		c = cosphi * math.Pow(math.Tan(fortPi+.5*phi1), n) / n
		if math.Abs(math.Abs(p.phi0)-halfPi) < epsilon10 {
			rho0 = 0.
		} else {
			rho0 = c * math.Pow(math.Tan(fortPi+.5*p.phi0), -n)
		}
	}
	p.fwd = func(lam float64, phi float64) (float64, float64, int) {
		var rho float64
		if math.Abs(math.Abs(phi)-halfPi) < epsilon10 {
			if phi*n <= 0. {
				return 0, 0, errOutsideProjectionDomain
			}
			rho = 0.
		} else if ell.es != 0. {
			rho = c * math.Pow(tsfn(phi, math.Sin(phi), ell.e), n)
		} else {
			rho = c * math.Pow(math.Tan(fortPi+.5*phi), -n)
		}
		lam *= n
		return p.k0 * (rho * math.Sin(lam)), p.k0 * (rho0 - rho*math.Cos(lam)), 0
	}
	p.inv = func(x float64, y float64) (float64, float64, int) {
		x /= p.k0
		y /= p.k0
		y = rho0 - y
		rho := math.Hypot(x, y)
		if rho == 0. {
			if n > 0. {
				return 0., halfPi, 0
			}
			return 0., -halfPi, 0
		}
		if n < 0. {
			rho = -rho
			x = -x
			y = -y
		}
		var phi float64
		if ell.es != 0. {
			phi = pjPhi2(math.Pow(rho/c, 1./n), ell.e)
			if math.IsInf(phi, 0) {
				return 0, 0, errOutsideProjectionDomain
			}
		} else {
			phi = 2.*math.Atan(math.Pow(c/rho, 1./n)) - halfPi
		}
		return math.Atan2(x, y) / n, phi, 0
	}
	return nil
}

// setupMerc sets up PROJ's Mercator.
func setupMerc(p *projection, params parameters) error {
	if params.has("lat_ts") {
		phits, err := params.angle("lat_ts", 0)
		if err != nil {
			return err
		}
		phits = math.Abs(phits)
		if phits >= halfPi {
			return errors.New("invalid value for lat_ts: |lat_ts| should be <= 90 degrees")
		}
		if p.ell.es != 0.0 {
			p.k0 = msfn(math.Sin(phits), math.Cos(phits), p.ell.es)
		} else {
			p.k0 = math.Cos(phits)
		}
	}
	if p.ell.es != 0.0 {
		e := p.ell.e
		p.fwd = func(lam float64, phi float64) (float64, float64, int) {
			sphi, cphi := math.Sin(phi), math.Cos(phi)
			return p.k0 * lam, p.k0 * (math.Asinh(sphi/cphi) - e*math.Atanh(e*sphi)), 0
		}
		p.inv = func(x float64, y float64) (float64, float64, int) {
			return x / p.k0, math.Atan(sinhpsi2tanphi(math.Sinh(y/p.k0), e)), 0
		}
		return nil
	}
	setupSphericalMerc(p)
	return nil
}

// setupWebMerc sets up PROJ's Web Mercator, which is the spherical Mercator on the ellipsoid's semimajor axis.
func setupWebMerc(p *projection, _ parameters) error {
	p.k0 = 1.0
	setupSphericalMerc(p)
	return nil
}

// setupSphericalMerc sets up PROJ's spherical Mercator.
func setupSphericalMerc(p *projection) {
	p.fwd = func(lam float64, phi float64) (float64, float64, int) {
		return p.k0 * lam, p.k0 * math.Asinh(math.Tan(phi)), 0
	}
	p.inv = func(x float64, y float64) (float64, float64, int) {
		return x / p.k0, math.Atan(math.Sinh(y / p.k0)), 0
	}
}

// authalicLatQ is PROJ's pj_authalic_lat_q.
func authalicLatQ(sinphi float64, ell ellipsoid) float64 {
	if ell.e >= 1e-7 {
		eSinphi := ell.e * sinphi
		oneMinusESinphiSq := 1.0 - eSinphi*eSinphi
		if oneMinusESinphiSq == 0.0 {
			return math.Inf(1)
		}
		return ell.oneEs * (sinphi/oneMinusESinphiSq + math.Atanh(eSinphi)/ell.e)
	}
	return 2 * sinphi
}

// authalicLatInverse is PROJ's pj_authalic_lat_inverse.
func authalicLatInverse(beta float64, apa []float64, ell ellipsoid, qp float64) float64 {
	phi := auxLatConvert(beta, apa[:auxLatOrder])
	if math.Abs(ell.n) < authalicSize {
		return phi
	}
	q := math.Sin(beta) * qp
	qDivOneMinusEs := q / ell.oneEs
	for i := 0; i < 10; i++ {
		sinphi, cosphi := math.Sin(phi), math.Cos(phi)
		oneMinusEsSin2phi := 1 - ell.es*(sinphi*sinphi)
		dphi := (oneMinusEsSin2phi * oneMinusEsSin2phi) / (2 * cosphi) *
			(qDivOneMinusEs - sinphi/oneMinusEsSin2phi - math.Atanh(ell.e*sinphi)/ell.e)
		if !(math.Abs(dphi) >= 1e-15) {
			break
		}
		phi += dphi
	}
	return phi
}

// setupAEA sets up PROJ's Albers Equal Area.
func setupAEA(p *projection, params parameters) error {
	phi1, err := params.angle("lat_1", 0)
	if err != nil {
		return err
	}
	phi2, err := params.angle("lat_2", 0)
	if err != nil {
		return err
	}
	if math.Abs(phi1) > halfPi || math.Abs(phi2) > halfPi {
		return errors.New("invalid value for lat_1 or lat_2: it should be within 90 degrees")
	}
	if math.Abs(phi1+phi2) < epsilon10 {
		return errors.New("invalid value for lat_1 and lat_2: |lat_1 + lat_2| should be > 0")
	}
	ell := p.ell
	sinphi := math.Sin(phi1)
	n := sinphi
	cosphi := math.Cos(phi1)
	secant := math.Abs(phi1-phi2) >= epsilon10
	ellips := ell.es > 0.
	var ec, c, dd, n2, rho0, qp float64
	var apa []float64
	if ellips {
		apa = auxLatCoeffs(ell.n, auxAuthalic, auxGeographic)
		qp = authalicLatQ(1.0, ell)
		m1 := msfn(sinphi, cosphi, ell.es)
		ml1 := authalicLatQ(sinphi, ell)
		if secant {
			sinphi = math.Sin(phi2)
			cosphi = math.Cos(phi2)
			m2 := msfn(sinphi, cosphi, ell.es)
			ml2 := authalicLatQ(sinphi, ell)
			if ml2 == ml1 {
				return errors.New("invalid value for lat_1 and lat_2")
			}
			n = (m1*m1 - m2*m2) / (ml2 - ml1)
			if n == 0 {
				return errors.New("invalid value for eccentricity")
			}
		}
		ec = 1. - .5*ell.oneEs*math.Log((1.-ell.e)/(1.+ell.e))/ell.e
		c = m1*m1 + n*ml1
		dd = 1. / n
		rho0 = dd * math.Sqrt(c-n*authalicLatQ(math.Sin(p.phi0), ell))
	} else {
		if secant {
			n = .5 * (n + math.Sin(phi2))
		}
		n2 = n + n
		c = cosphi*cosphi + n2*sinphi
		dd = 1. / n
		rho0 = dd * math.Sqrt(c-n2*math.Sin(p.phi0))
	}
	p.fwd = func(lam float64, phi float64) (float64, float64, int) {
		var rho float64
		if ellips {
			rho = c - n*authalicLatQ(math.Sin(phi), ell)
		} else {
			rho = c - n2*math.Sin(phi)
		}
		if rho < 0. {
			return 0, 0, errOutsideProjectionDomain
		}
		rho = dd * math.Sqrt(rho)
		lam *= n
		return rho * math.Sin(lam), rho0 - rho*math.Cos(lam), 0
	}
	p.inv = func(x float64, y float64) (float64, float64, int) {
		y = rho0 - y
		rho := math.Hypot(x, y)
		if rho == 0.0 {
			if n > 0. {
				return 0., halfPi, 0
			}
			return 0., -halfPi, 0
		}
		if n < 0. {
			rho = -rho
			x = -x
			y = -y
		}
		phi := rho / dd
		if ellips {
			qs := (c - phi*phi) / n
			if math.Abs(ec-math.Abs(qs)) > tolerance7 {
				if math.Abs(qs) > 2 {
					return 0, 0, errOutsideProjectionDomain
				}
				phi = authalicLatInverse(math.Asin(qs/qp), apa, ell, qp)
				if math.IsInf(phi, 0) {
					return 0, 0, errOutsideProjectionDomain
				}
			} else if qs < 0. {
				phi = -halfPi
			} else {
				phi = halfPi
			}
		} else {
			qsDiv2 := (c - phi*phi) / n2
			if math.Abs(qsDiv2) <= 1. {
				phi = math.Asin(qsDiv2)
			} else if qsDiv2 < 0. {
				phi = -halfPi
			} else {
				phi = halfPi
			}
		}
		return math.Atan2(x, y) / n, phi, 0
	}
	return nil
}

// cartesian is PROJ's cart conversion between geodetic and geocentric cartesian coordinates.
type cartesian struct {
	ell ellipsoid
}

var _ operation = cartesian{}

// units implements the interface operation.
func (cartesian) units() (ioUnits, ioUnits) {
	return unitsRadians, unitsCartesian
}

// normalRadiusOfCurvature is PROJ's normal_radius_of_curvature.
func normalRadiusOfCurvature(a float64, es float64, sinphi float64) float64 {
	if es == 0 {
		return a
	}
	return a / math.Sqrt(1-es*sinphi*sinphi)
}

// forward implements the interface operation.
func (o cartesian) forward(c *coordinate, _ *[4][]float64) int {
	if err := prepareAngular(c, 0, 0, false); err != 0 || c.failed() {
		return err
	}
	cosphi, sinphi := math.Cos(c[1]), math.Sin(c[1])
	n := normalRadiusOfCurvature(o.ell.a, o.ell.es, sinphi)
	lam, h := c[0], c[2]
	c[0] = (n + h) * cosphi * math.Cos(lam)
	c[1] = (n + h) * cosphi * math.Sin(lam)
	c[2] = (n*(1-o.ell.es) + h) * sinphi
	return 0
}

// inverse implements the interface operation.
func (o cartesian) inverse(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return errOutsideProjectionDomain
	}
	ell := o.ell
	xDivA, yDivA, zDivA := c[0]*ell.ra, c[1]*ell.ra, c[2]*ell.ra
	pDivA := math.Sqrt(xDivA*xDivA + yDivA*yDivA)
	bDivA := 1 - ell.f
	pDivABDivA := pDivA * bDivA
	norm := math.Sqrt(zDivA*zDivA + pDivABDivA*pDivABDivA)
	cc, s := 1.0, 0.0
	if norm != 0 {
		invNorm := 1.0 / norm
		cc = pDivABDivA * invNorm
		s = zDivA * invNorm
	}
	yPhi := zDivA + ell.e2s*bDivA*s*s*s
	xPhi := pDivA - ell.es*cc*cc*cc
	normPhi := math.Sqrt(yPhi*yPhi + xPhi*xPhi)
	cosphi, sinphi := 1.0, 0.0
	if normPhi != 0 {
		invNormPhi := 1.0 / normPhi
		cosphi = xPhi * invNormPhi
		sinphi = yPhi * invNormPhi
	}
	var phi float64
	if xPhi <= 0 {
		phi = halfPi
		sinphi = 1
		if c[2] < 0 {
			phi = -halfPi
			sinphi = -1
		}
		cosphi = 0
	} else {
		phi = math.Atan(yPhi / xPhi)
	}
	lam := math.Atan2(yDivA, xDivA)
	var h float64
	if cosphi < 1e-6 {
		cosphiSquared := cosphi * cosphi
		sinphiSquared := sinphi * sinphi
		bDivASquared := bDivA * bDivA
		bDivASquaredMulSinphiSquared := bDivASquared * sinphiSquared
		r := ell.a * math.Sqrt((cosphiSquared+bDivASquared*bDivASquaredMulSinphiSquared)/
			(cosphiSquared+bDivASquaredMulSinphiSquared))
		h = math.Abs(c[2]) - r
	} else {
		h = ell.a*pDivA/cosphi - normalRadiusOfCurvature(ell.a, ell.es, sinphi)
	}
	c[0], c[1], c[2] = lam, phi, h
	return finishAngular(c, 0, 0, false)
}

// helmert is PROJ's Helmert transformation of geocentric cartesian coordinates, without time-dependent parameters.
type helmert struct {
	x, y, z    float64
	scale      float64
	noRotation bool
	r          [3][3]float64
}

var _ operation = (*helmert)(nil)

// newHelmert returns the Helmert transformation that the parameters define.
func newHelmert(p parameters) (operation, error) {
	for _, unsupported := range []string{"dx", "dy", "dz", "drx", "dry", "drz", "ds", "t_epoch", "t_obs", "theta",
		"x_0", "y_0", "z_0", "px", "py", "pz"} {
		if p.has(unsupported) {
			return nil, errors.Errorf("unsupported Helmert parameter: %s", unsupported)
		}
	}
	h := &helmert{}
	var err error
	if h.x, err = p.float("x", 0); err != nil {
		return nil, err
	}
	if h.y, err = p.float("y", 0); err != nil {
		return nil, err
	}
	if h.z, err = p.float("z", 0); err != nil {
		return nil, err
	}
	var o, ph, k float64
	if o, err = p.float("rx", 0); err != nil {
		return nil, err
	}
	if ph, err = p.float("ry", 0); err != nil {
		return nil, err
	}
	if k, err = p.float("rz", 0); err != nil {
		return nil, err
	}
	if h.scale, err = p.float("s", 0); err != nil {
		return nil, err
	}
	o, ph, k = o*arcsecToRad, ph*arcsecToRad, k*arcsecToRad
	h.noRotation = o == 0 && ph == 0 && k == 0
	convention := p["convention"]
	if !h.noRotation && convention != "position_vector" && convention != "coordinate_frame" {
		return nil, errors.New("missing convention for the Helmert transformation")
	}
	if p.has("exact") {
		cf, sf := math.Cos(o), math.Sin(o)
		ct, st := math.Cos(ph), math.Sin(ph)
		cp, sp := math.Cos(k), math.Sin(k)
		h.r = [3][3]float64{
			{ct * cp, cf*sp + sf*st*cp, sf*sp - cf*st*cp},
			{-ct * sp, cf*cp - sf*st*sp, sf*cp + cf*st*sp},
			{st, -sf * ct, cf * ct},
		}
	} else {
		h.r = [3][3]float64{{1, k, -ph}, {-k, 1, o}, {ph, -o, 1}}
	}
	if convention == "position_vector" {
		h.r[0][1], h.r[1][0] = h.r[1][0], h.r[0][1]
		h.r[0][2], h.r[2][0] = h.r[2][0], h.r[0][2]
		h.r[1][2], h.r[2][1] = h.r[2][1], h.r[1][2]
	}
	return h, nil
}

// units implements the interface operation.
func (*helmert) units() (ioUnits, ioUnits) {
	return unitsCartesian, unitsCartesian
}

// forward implements the interface operation.
func (h *helmert) forward(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return 0
	}
	if h.noRotation && h.scale == 0 {
		c[0] += h.x
		c[1] += h.y
		c[2] += h.z
		return 0
	}
	scale := 1 + h.scale*1e-6
	x, y, z := c[0], c[1], c[2]
	r := &h.r
	c[0] = scale*(r[0][0]*x+r[0][1]*y+r[0][2]*z) + h.x
	c[1] = scale*(r[1][0]*x+r[1][1]*y+r[1][2]*z) + h.y
	c[2] = scale*(r[2][0]*x+r[2][1]*y+r[2][2]*z) + h.z
	return 0
}

// inverse implements the interface operation.
func (h *helmert) inverse(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return errOutsideProjectionDomain
	}
	if h.noRotation && h.scale == 0 {
		c[0] -= h.x
		c[1] -= h.y
		c[2] -= h.z
		return 0
	}
	scale := 1 + h.scale*1e-6
	x, y, z := (c[0]-h.x)/scale, (c[1]-h.y)/scale, (c[2]-h.z)/scale
	r := &h.r
	c[0] = r[0][0]*x + r[1][0]*y + r[2][0]*z
	c[1] = r[0][1]*x + r[1][1]*y + r[2][1]*z
	c[2] = r[0][2]*x + r[1][2]*y + r[2][2]*z
	return 0
}

// linearUnits holds PROJ's linear units and their sizes in meters.
var linearUnits = map[string]float64{
	"km": 1000.0, "m": 1.0, "dm": 0.1, "cm": 0.01, "mm": 0.001, "kmi": 1852.0, "in": 0.0254, "ft": 0.3048,
	"yd": 0.9144, "mi": 1609.344, "fath": 1.8288, "ch": 20.1168, "link": 0.201168, "us-in": 100 / 3937.0,
	"us-ft": 1200 / 3937.0, "us-yd": 3600 / 3937.0, "us-ch": 79200 / 3937.0, "us-mi": 6336000 / 3937.0,
	"ind-yd": 0.91439523, "ind-ft": 0.30479841, "ind-ch": 20.11669506,
}

// angularUnits holds PROJ's angular units and their sizes in radians.
var angularUnits = map[string]float64{"rad": 1.0, "deg": degToRad, "grad": math.Pi / 200}

// unitConvert is PROJ's unitconvert conversion of horizontal and vertical units.
type unitConvert struct {
	xyFactor, zFactor float64
	left, right       ioUnits
}

var _ operation = (*unitConvert)(nil)

// newUnitConvert returns the unit conversion that the parameters define.
func newUnitConvert(p parameters) (operation, error) {
	u := &unitConvert{xyFactor: 1, zFactor: 1}
	factor := func(name string) (float64, ioUnits, bool, error) {
		value, ok := p[name]
		if !ok {
			return 1, unitsWhatever, false, nil
		}
		if f, ok := linearUnits[value]; ok {
			return f, unitsWhatever, true, nil
		}
		if f, ok := angularUnits[value]; ok {
			units := unitsWhatever
			switch value {
			case "rad":
				units = unitsRadians
			case "deg":
				units = unitsDegrees
			}
			return f, units, false, nil
		}
		f, err := parseNumber(value)
		if err != nil || f == 0 || math.IsInf(1/f, 0) {
			return 0, unitsWhatever, false, errors.Errorf("unknown %s unit: %s", name, value)
		}
		return f, unitsWhatever, false, nil
	}
	if p.has("t_in") || p.has("t_out") {
		return nil, errors.New("time units are not supported")
	}
	in, left, _, err := factor("xy_in")
	if err != nil {
		return nil, err
	}
	out, right, _, err := factor("xy_out")
	if err != nil {
		return nil, err
	}
	u.xyFactor, u.left, u.right = in/out, left, right
	zIn, _, _, err := factor("z_in")
	if err != nil {
		return nil, err
	}
	zOut, _, _, err := factor("z_out")
	if err != nil {
		return nil, err
	}
	u.zFactor = zIn / zOut
	return u, nil
}

// units implements the interface operation.
func (u *unitConvert) units() (ioUnits, ioUnits) {
	return u.left, u.right
}

// forward implements the interface operation.
func (u *unitConvert) forward(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return 0
	}
	c[0] *= u.xyFactor
	c[1] *= u.xyFactor
	c[2] *= u.zFactor
	return 0
}

// inverse implements the interface operation.
func (u *unitConvert) inverse(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return 0
	}
	c[0] /= u.xyFactor
	c[1] /= u.xyFactor
	c[2] /= u.zFactor
	return 0
}

// axisSwap is PROJ's axisswap conversion, which reorders and negates axes.
type axisSwap struct {
	axis [4]int
	sign [4]float64
	n    int
}

var _ operation = (*axisSwap)(nil)

// newAxisSwap returns the axis swap that the parameters define.
func newAxisSwap(p parameters) (operation, error) {
	order, ok := p["order"]
	if !ok {
		return nil, errors.New("axisswap needs an order")
	}
	s := &axisSwap{}
	for i := range s.axis {
		s.axis[i], s.sign[i] = i, 1
	}
	parts := strings.Split(order, ",")
	if len(parts) > 4 {
		return nil, errors.New("invalid axisswap order")
	}
	for i, part := range parts {
		v, err := strconv.Atoi(part)
		if err != nil || v == 0 || v > 4 || v < -4 {
			return nil, errors.New("invalid axisswap order")
		}
		if v < 0 {
			s.sign[i] = -1
			v = -v
		}
		s.axis[i] = v - 1
	}
	s.n = len(parts)
	return s, nil
}

// units implements the interface operation.
func (s *axisSwap) units() (ioUnits, ioUnits) {
	return unitsWhatever, unitsWhatever
}

// forward implements the interface operation.
func (s *axisSwap) forward(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		return 0
	}
	in := *c
	for i := 0; i < s.n; i++ {
		c[i] = s.sign[i] * in[s.axis[i]]
	}
	return 0
}

// inverse implements the interface operation.
func (s *axisSwap) inverse(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		return 0
	}
	in := *c
	for i := 0; i < s.n; i++ {
		c[s.axis[i]] = s.sign[i] * in[i]
	}
	return 0
}

// pushPop is PROJ's push and pop conversions, which save ordinates on a stack and restore them.
type pushPop struct {
	push bool
	v    [4]bool
}

var _ operation = (*pushPop)(nil)

// units implements the interface operation.
func (*pushPop) units() (ioUnits, ioUnits) {
	return unitsWhatever, unitsWhatever
}

// save pushes the selected ordinates onto the stack.
func (o *pushPop) save(c *coordinate, stack *[4][]float64) {
	for i, selected := range o.v {
		if selected {
			stack[i] = append(stack[i], c[i])
		}
	}
}

// restore pops the selected ordinates from the stack.
func (o *pushPop) restore(c *coordinate, stack *[4][]float64) {
	for i, selected := range o.v {
		if selected && len(stack[i]) > 0 {
			c[i] = stack[i][len(stack[i])-1]
			stack[i] = stack[i][:len(stack[i])-1]
		}
	}
}

// forward implements the interface operation.
func (o *pushPop) forward(c *coordinate, stack *[4][]float64) int {
	if o.push {
		o.save(c, stack)
	} else {
		o.restore(c, stack)
	}
	return 0
}

// inverse implements the interface operation.
func (o *pushPop) inverse(c *coordinate, stack *[4][]float64) int {
	if o.push {
		o.restore(c, stack)
	} else {
		o.save(c, stack)
	}
	return 0
}

// geogOffset is PROJ's geogoffset transformation, which offsets longitudes, latitudes, and heights.
type geogOffset struct {
	dlon, dlat, dh float64
}

var _ operation = (*geogOffset)(nil)

// units implements the interface operation.
func (*geogOffset) units() (ioUnits, ioUnits) {
	return unitsRadians, unitsRadians
}

// forward implements the interface operation.
func (o *geogOffset) forward(c *coordinate, _ *[4][]float64) int {
	if err := prepareAngular(c, 0, 0, false); err != 0 || c.failed() {
		return err
	}
	c[0] += o.dlon
	c[1] += o.dlat
	c[2] += o.dh
	return 0
}

// inverse implements the interface operation.
func (o *geogOffset) inverse(c *coordinate, _ *[4][]float64) int {
	if c.failed() {
		c.fail()
		return errOutsideProjectionDomain
	}
	c[0] -= o.dlon
	c[1] -= o.dlat
	c[2] -= o.dh
	return finishAngular(c, 0, 0, false)
}

// noop is PROJ's noop conversion.
type noop struct{}

var _ operation = noop{}

// units implements the interface operation.
func (noop) units() (ioUnits, ioUnits) {
	return unitsWhatever, unitsWhatever
}

// forward implements the interface operation.
func (noop) forward(*coordinate, *[4][]float64) int {
	return 0
}

// inverse implements the interface operation.
func (noop) inverse(*coordinate, *[4][]float64) int {
	return 0
}

// newOperation returns the operation that the parameters define, along with whether it is a datum shift.
func newOperation(p parameters) (operation, bool, error) {
	name := p["proj"]
	switch name {
	case "noop":
		return noop{}, false, nil
	case "unitconvert":
		o, err := newUnitConvert(p)
		return o, false, err
	case "axisswap":
		o, err := newAxisSwap(p)
		return o, false, err
	case "push", "pop":
		o := &pushPop{push: name == "push"}
		for i := range o.v {
			o.v[i] = p.has("v_" + strconv.Itoa(i+1))
		}
		return o, true, nil
	case "cart":
		ell, err := newEllipsoid(p)
		return cartesian{ell: ell}, true, err
	case "helmert":
		o, err := newHelmert(p)
		return o, true, err
	case "geogoffset":
		o := &geogOffset{}
		var err error
		if o.dlon, err = p.float("dlon", 0); err != nil {
			return nil, true, err
		}
		if o.dlat, err = p.float("dlat", 0); err != nil {
			return nil, true, err
		}
		if o.dh, err = p.float("dh", 0); err != nil {
			return nil, true, err
		}
		o.dlon *= arcsecToRad
		o.dlat *= arcsecToRad
		return o, true, nil
	}
	o, ok, err := newProjection(name, p)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, errors.Wrapf(ErrUnsupportedTransformation, "the %s operation is not supported", name)
	}
	return o, false, nil
}
