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
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
)

// ErrUnsupportedTransformation is returned for transformations that PROJ would perform but that this package does
// not, such as those whose datum shift PROJ only chooses while transforming, or those that use an operation that this
// package does not implement. Callers should not fall back to another definition of the same systems, since that
// would give a different result than PROJ.
var ErrUnsupportedTransformation = errors.New("this coordinate transformation is not supported")

// step is an operation of a pipeline, which runs in reverse when inverted.
type step struct {
	op       operation
	inverted bool
	datum    bool
	text     string
}

// pipeline is a sequence of operations, which transforms coordinates forward or in reverse.
type pipeline []step

// inverse returns the pipeline that undoes this one.
func (p pipeline) inverse() pipeline {
	inverse := make(pipeline, len(p))
	for i, s := range p {
		s.inverted = !s.inverted
		inverse[len(p)-1-i] = s
	}
	return inverse
}

// datumSteps returns the text of the pipeline's datum shift steps.
func (p pipeline) datumSteps() string {
	var parts []string
	for _, s := range p {
		if s.datum {
			if s.inverted {
				parts = append(parts, "inv "+s.text)
			} else {
				parts = append(parts, s.text)
			}
		}
	}
	return strings.Join(parts, " | ")
}

// run transforms the coordinate, returning PROJ's error number on failure.
func (p pipeline) run(c *coordinate) int {
	var stack [4][]float64
	for _, s := range p {
		var err int
		if s.inverted {
			err = s.op.inverse(c, &stack)
		} else {
			err = s.op.forward(c, &stack)
		}
		if err != 0 {
			c.fail()
			return err
		}
	}
	return 0
}

// units returns the units of the pipeline's input and output.
func (p pipeline) units() (ioUnits, ioUnits) {
	if len(p) == 0 {
		return unitsWhatever, unitsWhatever
	}
	left := func(s step) ioUnits {
		l, r := s.op.units()
		if s.inverted {
			return r
		}
		return l
	}
	right := func(s step) ioUnits {
		l, r := s.op.units()
		if s.inverted {
			return l
		}
		return r
	}
	return left(p[0]), right(p[len(p)-1])
}

// parseParameters parses the parameters of a PROJ string, ignoring a leading '+' on each.
func parseParameters(tokens []string) parameters {
	p := make(parameters, len(tokens))
	for _, token := range tokens {
		key, value, _ := strings.Cut(strings.TrimPrefix(token, "+"), "=")
		if _, ok := p[key]; !ok {
			p[key] = value
		}
	}
	return p
}

// parsePipeline parses a PROJ string of a single operation or of a pipeline of steps. Parameters that a pipeline
// gives before its first step apply to every step.
func parsePipeline(definition string) (pipeline, error) {
	tokens := strings.Fields(definition)
	if len(tokens) == 0 {
		return nil, errors.New("empty PROJ string")
	}
	first := parseParameters(tokens[:1])
	if first["proj"] != "pipeline" {
		op, datum, err := newOperation(parseParameters(tokens))
		if err != nil {
			return nil, err
		}
		inverted := parseParameters(tokens).has("inv")
		return pipeline{{op: op, inverted: inverted, datum: datum, text: definition}}, nil
	}
	var groups [][]string
	var global []string
	for _, token := range tokens[1:] {
		switch strings.TrimPrefix(token, "+") {
		case "step":
			groups = append(groups, nil)
		default:
			if len(groups) == 0 {
				global = append(global, token)
			} else {
				groups[len(groups)-1] = append(groups[len(groups)-1], token)
			}
		}
	}
	var result pipeline
	for _, group := range groups {
		params := parseParameters(append(group, global...))
		for _, unsupported := range []string{"omit_fwd", "omit_inv", "omit_init"} {
			if params.has(unsupported) {
				return nil, errors.Wrapf(ErrUnsupportedTransformation, "the %s flag is not supported", unsupported)
			}
		}
		op, datum, err := newOperation(params)
		if err != nil {
			return nil, err
		}
		result = append(result, step{op: op, inverted: params.has("inv"), datum: datum, text: strings.Join(group, " ")})
	}
	return result, nil
}

// systemPipeline returns the pipeline that transforms longitudes and latitudes in degrees on WGS 84 into the given
// coordinate reference system, which may be named by an authority and code, a URN or URL naming one, or a PROJ
// string. It returns ErrUnsupportedTransformation for systems that PROJ supports but this package does not, and
// another error for definitions that PROJ cannot read, or that PROJ reads in ways that this package does not, such
// as WKT.
func systemPipeline(definition string) (pipeline, error) {
	definition = strings.TrimLeft(definition, " \t\n\r")
	if strings.HasPrefix(definition, "+") || strings.HasPrefix(definition, "proj=") {
		return projStringPipeline(definition)
	}
	if strings.Contains(definition, "[") {
		return nil, errors.New("WKT definitions are not supported")
	}
	entry, ok := lookUpDefinition(definition)
	if !ok {
		return nil, errors.Errorf("crs not found: %s", definition)
	}
	switch entry.pipeline {
	case "!RUNTIME":
		return nil, errors.Wrapf(ErrUnsupportedTransformation, "PROJ chooses the datum shift to %s:%s while transforming",
			entry.Authority, entry.Code)
	case "!NONE":
		return nil, errors.Errorf("PROJ has no operation to %s:%s", entry.Authority, entry.Code)
	}
	p, err := parsePipeline(entry.pipeline)
	if err != nil {
		return nil, errors.Wrapf(ErrUnsupportedTransformation, "%s:%s: %s", entry.Authority, entry.Code, err.Error())
	}
	return p, nil
}

// lookUpDefinition returns the catalog's system that a definition names by its authority and code, or by a URN or
// URL naming them, following the same matching as readAuthorityCode.
func lookUpDefinition(definition string) (*crsEntry, bool) {
	authority, code, ok := urnAuthorityCode(definition)
	if !ok {
		if strings.Contains(definition, " ") {
			return nil, false
		}
		parts := strings.Split(definition, ":")
		if len(parts) != 2 {
			return nil, false
		}
		authority, code = parts[0], parts[1]
	}
	c := catalog()
	if entry, ok := c.lookUp(authority, code); ok {
		return entry, true
	}
	if strings.EqualFold(authority, "EPSG") && code == "102100" {
		if entry, ok := c.lookUp("ESRI", code); ok {
			return entry, true
		}
	}
	for _, name := range versionedAuthorities[authority] {
		if entry, ok := c.lookUp(name, code); ok {
			return entry, true
		}
	}
	for name := range c.byCode {
		if strings.EqualFold(name, authority) {
			if entry, ok := c.lookUp(name, code); ok {
				return entry, true
			}
		}
	}
	return nil, false
}

// projStringPipeline returns the pipeline from longitudes and latitudes in degrees on WGS 84 into the system that a
// PROJ string defines, as PROJ forms it: a datum shift through geocentric coordinates when the string gives a nonzero
// towgs84, either directly or through a named datum, and none otherwise, since PROJ then treats the datum as WGS 84.
func projStringPipeline(definition string) (pipeline, error) {
	params := parseParameters(strings.Fields(definition))
	if params.has("init") {
		return nil, errors.New("init files are not supported")
	}
	name, ok := params["proj"]
	if !ok || name == "" {
		return nil, errors.Errorf("could not parse proj string '%s'", definition)
	}
	if name == "pipeline" {
		return nil, errors.New("a pipeline is not a coordinate reference system")
	}
	for _, unsupported := range []string{"geoidgrids", "geoc", "vunits", "vto_meter", "lon_wrap", "over"} {
		if params.has(unsupported) {
			return nil, errors.Wrapf(ErrUnsupportedTransformation, "the %s parameter is not supported", unsupported)
		}
	}
	towgs84 := params["towgs84"]
	nadgrids := params["nadgrids"]
	if datumName, ok := params["datum"]; ok {
		d, ok := datums[datumName]
		if !ok {
			return nil, errors.Errorf("unknown datum: %s", datumName)
		}
		if towgs84 == "" && nadgrids == "" {
			towgs84, nadgrids = d.towgs84, d.nadgrids
		}
	}
	if name == "lcc" && !params.has("lat_2") && !params.has("lat_0") {
		params["lat_2"] = "0"
	}
	if name == "geocent" {
		towgs84, nadgrids = "", ""
	}
	if nadgrids != "" && nadgrids != "@null" {
		return nil, errors.Wrapf(ErrUnsupportedTransformation, "datum shift grids are not supported")
	}
	if _, err := newEllipsoid(params); err != nil {
		return nil, err
	}
	local := ellipsoidText(params)
	shift, shifted, err := helmertFromTowgs84(towgs84)
	if err != nil {
		return nil, err
	}
	steps := []string{"proj=unitconvert xy_in=deg xy_out=rad"}
	if shifted {
		steps = append(steps, "proj=push v_3", "proj=cart ellps=WGS84", "inv "+shift, "inv proj=cart "+local, "proj=pop v_3")
	}
	var system []string
	for key, value := range params {
		switch key {
		case "proj", "type", "no_defs", "towgs84", "nadgrids", "datum", "wktext", "units", "to_meter", "ellps", "a",
			"b", "rf", "f", "es", "e", "R":
			continue
		}
		if value == "" {
			system = append(system, key)
		} else {
			system = append(system, key+"="+value)
		}
	}
	sort.Strings(system)
	system = append(system, local)
	switch name {
	case "longlat", "latlong", "lonlat", "latlon":
		if !shifted {
			return pipeline{{op: noop{}, text: "proj=noop"}}, nil
		}
		steps = append(steps, "proj=unitconvert xy_in=rad xy_out=deg")
	case "geocent":
		steps = append(steps, "proj=cart "+local)
	default:
		steps = append(steps, "proj="+name+" "+strings.Join(system, " "))
		if units, ok := params["units"]; ok && units != "m" {
			steps = append(steps, "proj=unitconvert xy_in=m xy_out="+units)
		} else if toMeter, ok := params["to_meter"]; ok {
			steps = append(steps, "proj=unitconvert xy_in=m xy_out="+toMeter)
		}
	}
	p, err := parsePipeline("proj=pipeline step " + strings.Join(steps, " step "))
	if err != nil {
		if errors.Is(err, ErrUnsupportedTransformation) {
			return nil, err
		}
		return nil, errors.Errorf("could not parse proj string '%s': %s", definition, err.Error())
	}
	return p, nil
}

// helmertFromTowgs84 returns the Helmert step of a towgs84 shift, which uses the position vector convention, and
// false when the shift is absent or zero.
func helmertFromTowgs84(towgs84 string) (string, bool, error) {
	if towgs84 == "" {
		return "", false, nil
	}
	values := strings.Split(towgs84, ",")
	if len(values) != 3 && len(values) != 7 {
		return "", false, errors.Errorf("invalid towgs84: %s", towgs84)
	}
	zero := true
	for _, v := range values {
		f, err := parseNumber(v)
		if err != nil {
			return "", false, errors.Errorf("invalid towgs84: %s", towgs84)
		}
		zero = zero && f == 0
	}
	if zero {
		return "", false, nil
	}
	shift := "proj=helmert x=" + values[0] + " y=" + values[1] + " z=" + values[2]
	if len(values) == 7 {
		shift += " rx=" + values[3] + " ry=" + values[4] + " rz=" + values[5] + " s=" + values[6] +
			" convention=position_vector"
	}
	return shift, true, nil
}

// ellipsoidText returns the parameters of a PROJ string that define its ellipsoid, taking a named datum's ellipsoid
// when the string names no other.
func ellipsoidText(params parameters) string {
	var parts []string
	for _, key := range []string{"ellps", "R", "a", "b", "rf", "f", "es", "e"} {
		if value, ok := params[key]; ok {
			parts = append(parts, key+"="+value)
		}
	}
	if len(parts) == 0 {
		if d, ok := datums[params["datum"]]; ok {
			return "ellps=" + d.ellipsoid
		}
		return "ellps=GRS80"
	}
	return strings.Join(parts, " ")
}

// Transformer transforms coordinates from one coordinate reference system to another. Geographic coordinates are in
// degrees with the longitude first. A transformer may only be used by one goroutine at a time.
type Transformer struct {
	pipeline      pipeline
	fails         bool
	angularInput  bool
	angularOutput bool
}

// NewTransformer returns a transformer between the given coordinate reference systems, each given as an authority
// and code, a URN or URL naming one, or a PROJ string. Pairs of systems whose datum shifts both differ from WGS 84
// return ErrUnsupportedTransformation, since PROJ may shift between them directly.
func NewTransformer(source string, target string) (*Transformer, error) {
	from, err := systemPipeline(source)
	if err != nil {
		return nil, err
	}
	to, err := systemPipeline(target)
	if err != nil {
		return nil, err
	}
	fromDatum, toDatum := from.datumSteps(), to.datumSteps()
	if fromDatum != "" && toDatum != "" && fromDatum != toDatum {
		return nil, errors.Wrap(ErrUnsupportedTransformation, "both systems shift their datums from WGS 84")
	}
	return &Transformer{pipeline: append(from.inverse(), to...)}, nil
}

// NewPipelineTransformer returns a transformer that runs the given coordinate operation forward or in reverse, which
// is a PROJ string of an operation or of a pipeline. Like PROJ, it reads a coordinate reference system named by its
// authority and code, but transforming with it fails. Operations whose input or output is in radians take and give
// degrees.
func NewPipelineTransformer(definition string, forward bool) (*Transformer, error) {
	trimmed := strings.TrimLeft(definition, " \t\n\r")
	if !strings.HasPrefix(trimmed, "+") && !strings.HasPrefix(trimmed, "proj=") {
		if readDefinition(trimmed) != unreadable {
			return &Transformer{fails: true}, nil
		}
		return nil, errors.Errorf("could not parse coordinate operation '%s'", definition)
	}
	p, err := parsePipeline(trimmed)
	if err != nil {
		return nil, err
	}
	if !forward {
		p = p.inverse()
	}
	left, right := p.units()
	return &Transformer{pipeline: p, angularInput: left == unitsRadians, angularOutput: right == unitsRadians}, nil
}

// Close releases the transformer.
func (t *Transformer) Close() {}

// Transform transforms the interleaved coordinates in place, which hold `dims` ordinates per point. The third ordinate
// of each point is transformed as a height when `hasZ` is true, and any other ordinates are left alone. The error
// carries PROJ's error number.
func (t *Transformer) Transform(coords []float64, dims int, hasZ bool) error {
	count := len(coords) / dims
	if count == 0 {
		return nil
	}
	if t.fails {
		return TransformError{Message: "Invalid coordinate", Number: errInvalidCoordinate}
	}
	lastErr := 0
	for i := 0; i < count; i++ {
		point := coords[i*dims : i*dims+dims]
		c := coordinate{point[0], point[1], 0, 0}
		if hasZ {
			c[2] = point[2]
		}
		if t.angularInput {
			c[0] *= math.Pi / 180
			c[1] *= math.Pi / 180
		}
		if err := t.pipeline.run(&c); err != 0 {
			lastErr = err
		}
		if t.angularOutput && !c.failed() {
			c[0] *= 180 / math.Pi
			c[1] *= 180 / math.Pi
		}
		point[0], point[1] = c[0], c[1]
		if hasZ {
			point[2] = c[2]
		}
	}
	if lastErr != 0 {
		message, ok := errorMessages[lastErr]
		if !ok {
			message = "Unknown error"
		}
		return TransformError{Message: message, Number: lastErr}
	}
	return nil
}
