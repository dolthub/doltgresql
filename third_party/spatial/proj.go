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

/*
#include <stdlib.h>
#include <proj.h>

static const char *spatial_wkt_options[] = {"MULTILINE=NO", NULL};

static const char *spatial_as_wkt(PJ_CONTEXT *ctx, PJ *crs) {
	return proj_as_wkt(ctx, crs, PJ_WKT1_GDAL, spatial_wkt_options);
}
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/cockroachdb/errors"
)

// CRSKind is the kind of a coordinate reference system.
type CRSKind int

// The kinds of coordinate reference systems in PROJ's database that the PostGIS functions list.
const (
	OtherCRS CRSKind = iota
	GeocentricCRS
	Geographic2DCRS
	Geographic3DCRS
	ProjectedCRS
	CompoundCRS
)

// CRSInfo describes a coordinate reference system in PROJ's database, along with the longitude and latitude bounds of
// its area of use when it has one.
type CRSInfo struct {
	Authority  string
	Code       string
	Name       string
	Kind       CRSKind
	Deprecated bool
	AreaValid  bool
	West       float64
	South      float64
	East       float64
	North      float64
}

// newContext returns a new PROJ context that writes nothing to its log, as its errors reach callers through projError.
func newContext() *C.PJ_CONTEXT {
	ctx := C.proj_context_create()
	C.proj_log_level(ctx, C.PJ_LOG_NONE)
	return ctx
}

// DatabaseCRSInfo returns the coordinate reference systems of the given authority in PROJ's database. When `within`
// holds a west, south, east, and north bound, it returns only the projected systems of Earth that are not deprecated
// and whose area of use holds those bounds.
func DatabaseCRSInfo(authority string, within *[4]float64) ([]CRSInfo, error) {
	ctx := newContext()
	defer C.proj_context_destroy(ctx)
	cAuthority := C.CString(authority)
	defer C.free(unsafe.Pointer(cAuthority))
	var params *C.PROJ_CRS_LIST_PARAMETERS
	if within != nil {
		params = C.proj_get_crs_list_parameters_create()
		defer C.proj_get_crs_list_parameters_destroy(params)
		projected := (*C.PJ_TYPE)(C.malloc(C.size_t(unsafe.Sizeof(C.PJ_TYPE(0)))))
		defer C.free(unsafe.Pointer(projected))
		*projected = C.PJ_TYPE_PROJECTED_CRS
		params.types, params.typesCount = projected, 1
		params.crs_area_of_use_contains_bbox, params.bbox_valid, params.allow_deprecated = 1, 1, 0
		params.west_lon_degree, params.south_lat_degree = C.double(within[0]), C.double(within[1])
		params.east_lon_degree, params.north_lat_degree = C.double(within[2]), C.double(within[3])
		earth := C.CString("Earth")
		defer C.free(unsafe.Pointer(earth))
		params.celestial_body_name = earth
	}
	var count C.int
	list := C.proj_get_crs_info_list_from_database(ctx, cAuthority, params, &count)
	if list == nil {
		return nil, projError(ctx)
	}
	defer C.proj_crs_info_list_destroy(list)
	kinds := map[C.PJ_TYPE]CRSKind{
		C.PJ_TYPE_GEOCENTRIC_CRS:    GeocentricCRS,
		C.PJ_TYPE_GEOGRAPHIC_2D_CRS: Geographic2DCRS,
		C.PJ_TYPE_GEOGRAPHIC_3D_CRS: Geographic3DCRS,
		C.PJ_TYPE_PROJECTED_CRS:     ProjectedCRS,
		C.PJ_TYPE_COMPOUND_CRS:      CompoundCRS,
	}
	infos := make([]CRSInfo, 0, int(count))
	for _, info := range unsafe.Slice(list, int(count)) {
		infos = append(infos, CRSInfo{
			Authority:  C.GoString(info.auth_name),
			Code:       C.GoString(info.code),
			Name:       C.GoString(info.name),
			Kind:       kinds[info._type],
			Deprecated: info.deprecated != 0,
			AreaValid:  info.bbox_valid != 0,
			West:       float64(info.west_lon_degree),
			South:      float64(info.south_lat_degree),
			East:       float64(info.east_lon_degree),
			North:      float64(info.north_lat_degree),
		})
	}
	return infos, nil
}

// CRSDefinitions returns the WKT, as GDAL writes it, and the PROJ string of a coordinate reference system in PROJ's
// database. The PROJ string is empty when PROJ cannot express the system as one.
func CRSDefinitions(authority string, code string) (wkt string, projString string, err error) {
	ctx := newContext()
	defer C.proj_context_destroy(ctx)
	cAuthority, cCode := C.CString(authority), C.CString(code)
	defer C.free(unsafe.Pointer(cAuthority))
	defer C.free(unsafe.Pointer(cCode))
	crs := C.proj_create_from_database(ctx, cAuthority, cCode, C.PJ_CATEGORY_CRS, 0, nil)
	if crs == nil {
		return "", "", projError(ctx)
	}
	defer C.proj_destroy(crs)
	wkt = C.GoString(C.spatial_as_wkt(ctx, crs))
	if cProjString := C.proj_as_proj_string(ctx, crs, C.PJ_PROJ_4, nil); cProjString != nil {
		projString = C.GoString(cProjString)
	}
	return wkt, projString, nil
}

// Transformer transforms coordinates from one coordinate reference system to another. Geographic coordinates are in
// degrees with the longitude first. A transformer may only be used by one goroutine at a time.
type Transformer struct {
	ctx           *C.PJ_CONTEXT
	pj            *C.PJ
	direction     C.PJ_DIRECTION
	angularInput  bool
	angularOutput bool
}

// NewTransformer returns a transformer between the given coordinate reference systems, each given in any form that
// PROJ accepts, such as an authority and code, WKT, or a PROJ string. The transformer must be closed once it is no
// longer needed.
func NewTransformer(source string, target string) (*Transformer, error) {
	ctx := newContext()
	cSource, cTarget := C.CString(source), C.CString(target)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cTarget))
	pj := C.proj_create_crs_to_crs(ctx, cSource, cTarget, nil)
	if pj == nil {
		err := projError(ctx)
		C.proj_context_destroy(ctx)
		return nil, err
	}
	return newNormalizedTransformer(ctx, pj, C.PJ_FWD)
}

// NewPipelineTransformer returns a transformer that runs the given coordinate operation, which may be a PROJ pipeline
// or an operation from PROJ's database, forward or in reverse. The transformer must be closed once it is no longer
// needed.
func NewPipelineTransformer(pipeline string, forward bool) (*Transformer, error) {
	ctx := newContext()
	cPipeline := C.CString(pipeline)
	defer C.free(unsafe.Pointer(cPipeline))
	pj := C.proj_create(ctx, cPipeline)
	if pj == nil {
		err := projError(ctx)
		C.proj_context_destroy(ctx)
		return nil, err
	}
	direction := C.PJ_DIRECTION(C.PJ_FWD)
	if !forward {
		direction = C.PJ_INV
	}
	return newNormalizedTransformer(ctx, pj, direction)
}

// newNormalizedTransformer returns a transformer for the given operation in the given direction, after normalizing
// its axis order so that geographic coordinates take the longitude first. An operation without coordinate reference
// systems to normalize is used as it is, converting between degrees and the radians that it works in. It takes
// ownership of the context and the operation.
func newNormalizedTransformer(ctx *C.PJ_CONTEXT, pj *C.PJ, direction C.PJ_DIRECTION) (*Transformer, error) {
	if normalized := C.proj_normalize_for_visualization(ctx, pj); normalized != nil {
		C.proj_destroy(pj)
		pj = normalized
	}
	return &Transformer{
		ctx:           ctx,
		pj:            pj,
		direction:     direction,
		angularInput:  C.proj_angular_input(pj, int32(direction)) == 1,
		angularOutput: C.proj_angular_output(pj, int32(direction)) == 1,
	}, nil
}

// Reads returns whether PROJ can read the given definition.
func Reads(definition string) bool {
	ctx := newContext()
	defer C.proj_context_destroy(ctx)
	cDefinition := C.CString(definition)
	defer C.free(unsafe.Pointer(cDefinition))
	pj := C.proj_create(ctx, cDefinition)
	if pj == nil {
		return false
	}
	C.proj_destroy(pj)
	return true
}

// Close releases the transformer.
func (t *Transformer) Close() {
	C.proj_destroy(t.pj)
	C.proj_context_destroy(t.ctx)
}

// Transform transforms the interleaved coordinates in place, which hold `dims` ordinates per point. The third ordinate
// of each point is transformed as a height when `hasZ` is true, and any other ordinates are left alone. The error
// carries PROJ's error number.
func (t *Transformer) Transform(coords []float64, dims int, hasZ bool) error {
	count := len(coords) / dims
	if count == 0 {
		return nil
	}
	stride := C.size_t(dims * 8)
	x, y := (*C.double)(&coords[0]), (*C.double)(&coords[1])
	var z *C.double
	var zCount C.size_t
	if hasZ {
		z, zCount = (*C.double)(&coords[2]), C.size_t(count)
	}
	if t.angularInput {
		scaleXY(coords, dims, math.Pi/180)
	}
	C.proj_errno_reset(t.pj)
	C.proj_trans_generic(t.pj, t.direction, x, stride, C.size_t(count), y, stride, C.size_t(count), z, stride, zCount, nil, 0, 0)
	if errno := C.proj_errno(t.pj); errno != 0 {
		return TransformError{Message: C.GoString(C.proj_context_errno_string(t.ctx, errno)), Number: int(errno)}
	}
	if t.angularOutput {
		scaleXY(coords, dims, 180/math.Pi)
	}
	return nil
}

// scaleXY multiplies the first two ordinates of every point by the factor.
func scaleXY(coords []float64, dims int, factor float64) {
	for i := 0; i+1 < len(coords); i += dims {
		coords[i] *= factor
		coords[i+1] *= factor
	}
}

// TransformError is an error that PROJ reported while transforming coordinates.
type TransformError struct {
	Message string
	Number  int
}

// Error implements the interface error.
func (e TransformError) Error() string {
	return e.Message
}

// projError returns the last error that PROJ reported in the given context.
func projError(ctx *C.PJ_CONTEXT) error {
	return errors.New(C.GoString(C.proj_context_errno_string(ctx, C.proj_context_errno(ctx))))
}

// PROJSettings returns whether PROJ may download data over the network, the address it would download from, the
// directory it writes downloaded data to, and the location of its database.
func PROJSettings() (networkEnabled bool, urlEndpoint string, writableDirectory string, databasePath string) {
	ctx := newContext()
	defer C.proj_context_destroy(ctx)
	return C.proj_context_is_network_enabled(ctx) == 1, C.GoString(C.proj_context_get_url_endpoint(ctx)),
		C.GoString(C.proj_context_get_user_writable_directory(ctx, 0)), C.GoString(C.proj_context_get_database_path(ctx))
}

// IsGeographic returns whether PROJ reads the given definition as a geographic coordinate reference system.
func IsGeographic(definition string) bool {
	ctx := newContext()
	defer C.proj_context_destroy(ctx)
	cDefinition := C.CString(definition)
	defer C.free(unsafe.Pointer(cDefinition))
	pj := C.proj_create(ctx, cDefinition)
	if pj == nil {
		return false
	}
	defer C.proj_destroy(pj)
	kind := C.proj_get_type(pj)
	return kind == C.PJ_TYPE_GEOGRAPHIC_2D_CRS || kind == C.PJ_TYPE_GEOGRAPHIC_3D_CRS
}
