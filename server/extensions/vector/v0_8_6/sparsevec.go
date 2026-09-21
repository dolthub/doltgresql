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

package v0_8_6

import (
	"encoding/binary"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
)

const (
	// maxSparseDims is the dimension limit that pgvector enforces for the sparsevec type.
	maxSparseDims = 1000000000
	// maxSparseNonZero is the limit on non-zero elements that pgvector enforces for the sparsevec type.
	maxSparseNonZero = 16000
)

// sparseVector is the Go representation of a sparsevec value: the non-zero elements of a vector, with zero-based
// indices in ascending order.
type sparseVector struct {
	Dim     int32
	Indices []int32
	Values  []float32
}

// checkSparseDims returns an error unless both sparse vectors have the same number of dimensions.
func checkSparseDims(a sparseVector, b sparseVector) error {
	if a.Dim != b.Dim {
		return errors.Errorf("different sparsevec dimensions %d and %d", a.Dim, b.Dim)
	}
	return nil
}

// parseSparseInt parses an index or dimension count from a sparsevec literal, clamping out-of-range values into
// [minVal, math.MaxInt32] like pgvector so the bounds checks report the specific error.
func parseSparseInt(str string, minVal int64) (int64, error) {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		if numErr, ok := err.(*strconv.NumError); !ok || numErr.Err != strconv.ErrRange {
			return 0, err
		}
	}
	if val > math.MaxInt32 {
		val = math.MaxInt32
	} else if val < minVal {
		val = minVal
	}
	return val, nil
}

// hasNonZeroMantissa reports whether an element token's mantissa contains a non-zero digit, identifying values that
// only became zero through float32 underflow.
func hasNonZeroMantissa(token string) bool {
	for i := 0; i < len(token); i++ {
		ch := token[i]
		if ch == 'e' || ch == 'E' {
			return false
		}
		if ch >= '1' && ch <= '9' {
			return true
		}
	}
	return false
}

// sparseFromDense converts a dense vector to its sparse representation, keeping only the non-zero elements.
func sparseFromDense(vals []float32) sparseVector {
	sv := sparseVector{Dim: int32(len(vals))}
	for i, val := range vals {
		if val != 0 {
			sv.Indices = append(sv.Indices, int32(i))
			sv.Values = append(sv.Values, val)
		}
	}
	return sv
}

// sparseToDense converts a sparse vector to its dense representation.
func sparseToDense(sv sparseVector) []float32 {
	vals := make([]float32, sv.Dim)
	for i, index := range sv.Indices {
		vals[index] = sv.Values[i]
	}
	return vals
}

// sparsevecIn implements sparsevec_in, which parses the "{index:value,...}/dimensions" text representation.
func sparsevecIn(ctx *sql.Context, args ...any) (any, error) {
	input := args[0].(string)
	trimmed := strings.TrimSpace(input)
	closing := strings.IndexByte(trimmed, '}')
	if !strings.HasPrefix(trimmed, "{") || closing == -1 {
		return nil, errMalformedLiteral("sparsevec", input)
	}
	after := strings.TrimSpace(trimmed[closing+1:])
	if !strings.HasPrefix(after, "/") {
		return nil, errMalformedLiteral("sparsevec", input)
	}
	dim, err := parseSparseInt(strings.TrimSpace(after[1:]), math.MinInt32)
	if err != nil {
		return nil, errMalformedLiteral("sparsevec", input)
	}
	if dim < 1 {
		return nil, errAtLeastOneDim("sparsevec")
	}
	if dim > maxSparseDims {
		return nil, errTooManyDims("sparsevec", maxSparseDims)
	}
	type sparseEntry struct {
		index int32
		value float32
	}
	var entries []sparseEntry
	if inner := strings.TrimSpace(trimmed[1:closing]); len(inner) > 0 {
		for _, pair := range strings.Split(inner, ",") {
			colon := strings.IndexByte(pair, ':')
			if colon == -1 {
				return nil, errMalformedLiteral("sparsevec", input)
			}
			index, err := parseSparseInt(strings.TrimSpace(pair[:colon]), math.MinInt32+1)
			if err != nil {
				return nil, errMalformedLiteral("sparsevec", input)
			}
			if index < 1 || index > dim {
				return nil, errors.New("sparsevec index out of bounds")
			}
			token := strings.TrimSpace(pair[colon+1:])
			val, err := parseElement("sparsevec", input, token)
			if err != nil {
				return nil, err
			}
			if val == 0 && hasNonZeroMantissa(token) {
				return nil, errors.Errorf(`"%s" is out of range for type sparsevec`, token)
			}
			if err = checkElement("sparsevec", val); err != nil {
				return nil, err
			}
			entries = append(entries, sparseEntry{index: int32(index - 1), value: val})
		}
	}
	if len(entries) > maxSparseNonZero {
		return nil, errors.Errorf("sparsevec cannot have more than %d non-zero elements", maxSparseNonZero)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].index < entries[j].index
	})
	sv := sparseVector{Dim: int32(dim)}
	for i, entry := range entries {
		if i > 0 && entry.index == entries[i-1].index {
			return nil, errors.New("sparsevec indices must not contain duplicates")
		}
		if entry.value != 0 {
			sv.Indices = append(sv.Indices, entry.index)
			sv.Values = append(sv.Values, entry.value)
		}
	}
	if err = checkExpectedDims(args[2].(int32), int(sv.Dim)); err != nil {
		return nil, err
	}
	return sv, nil
}

// sparsevecOut implements sparsevec_out, which renders the "{index:value,...}/dimensions" text representation.
func sparsevecOut(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	sb := strings.Builder{}
	sb.WriteByte('{')
	for i, index := range sv.Indices {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatInt(int64(index)+1, 10))
		sb.WriteByte(':')
		sb.WriteString(formatElement(sv.Values[i]))
	}
	sb.WriteString("}/")
	sb.WriteString(strconv.FormatInt(int64(sv.Dim), 10))
	return sb.String(), nil
}

// sparsevecTypmodIn implements sparsevec_typmod_in, which parses the dimensions modifier.
func sparsevecTypmodIn(ctx *sql.Context, args ...any) (any, error) {
	return typmodInValue("sparsevec", maxSparseDims, args[0].([]any))
}

// sparsevecRecv implements sparsevec_recv, which decodes the binary representation of a sparse vector.
func sparsevecRecv(ctx *sql.Context, args ...any) (any, error) {
	data := args[0].([]byte)
	if len(data) < 12 {
		return nil, errors.New("insufficient data left in message")
	}
	dim := int32(binary.BigEndian.Uint32(data))
	nonZero := int32(binary.BigEndian.Uint32(data[4:]))
	unused := int32(binary.BigEndian.Uint32(data[8:]))
	if dim < 1 {
		return nil, errAtLeastOneDim("sparsevec")
	}
	if dim > maxSparseDims {
		return nil, errTooManyDims("sparsevec", maxSparseDims)
	}
	if nonZero < 0 || nonZero > maxSparseNonZero {
		return nil, errors.Errorf("sparsevec cannot have more than %d non-zero elements", maxSparseNonZero)
	}
	if unused != 0 {
		return nil, errors.Errorf("expected unused to be 0, not %d", unused)
	}
	if len(data) != 12+int(nonZero)*8 {
		return nil, errors.New("insufficient data left in message")
	}
	sv := sparseVector{
		Dim:     dim,
		Indices: make([]int32, nonZero),
		Values:  make([]float32, nonZero),
	}
	for i := range sv.Indices {
		sv.Indices[i] = int32(binary.BigEndian.Uint32(data[12+i*4:]))
	}
	valueData := data[12+int(nonZero)*4:]
	for i := range sv.Values {
		sv.Values[i] = math.Float32frombits(binary.BigEndian.Uint32(valueData[i*4:]))
		if err := checkElement("sparsevec", sv.Values[i]); err != nil {
			return nil, err
		}
	}
	if err := checkExpectedDims(args[2].(int32), int(sv.Dim)); err != nil {
		return nil, err
	}
	return sv, nil
}

// sparsevecSend implements sparsevec_send, which encodes the binary representation of a sparse vector.
func sparsevecSend(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	data := make([]byte, 12+len(sv.Indices)*8)
	binary.BigEndian.PutUint32(data, uint32(sv.Dim))
	binary.BigEndian.PutUint32(data[4:], uint32(len(sv.Indices)))
	for i, index := range sv.Indices {
		binary.BigEndian.PutUint32(data[12+i*4:], uint32(index))
	}
	valueData := data[12+len(sv.Indices)*4:]
	for i, val := range sv.Values {
		binary.BigEndian.PutUint32(valueData[i*4:], math.Float32bits(val))
	}
	return data, nil
}

// sparseL2Squared returns the squared Euclidean distance, merging the two vectors in index order and accumulating in
// float32 like pgvector.
func sparseL2Squared(a sparseVector, b sparseVector) float32 {
	distance := float32(0)
	i, j := 0, 0
	for i < len(a.Indices) && j < len(b.Indices) {
		switch {
		case a.Indices[i] == b.Indices[j]:
			diff := a.Values[i] - b.Values[j]
			distance += diff * diff
			i++
			j++
		case a.Indices[i] < b.Indices[j]:
			distance += a.Values[i] * a.Values[i]
			i++
		default:
			distance += b.Values[j] * b.Values[j]
			j++
		}
	}
	for ; i < len(a.Indices); i++ {
		distance += a.Values[i] * a.Values[i]
	}
	for ; j < len(b.Indices); j++ {
		distance += b.Values[j] * b.Values[j]
	}
	return distance
}

// sparseInnerProduct returns the inner product, which only the shared indices contribute to.
func sparseInnerProduct(a sparseVector, b sparseVector) float32 {
	product := float32(0)
	i, j := 0, 0
	for i < len(a.Indices) && j < len(b.Indices) {
		switch {
		case a.Indices[i] == b.Indices[j]:
			product += a.Values[i] * b.Values[j]
			i++
			j++
		case a.Indices[i] < b.Indices[j]:
			i++
		default:
			j++
		}
	}
	return product
}

// sparsevecL2Distance implements l2_distance for sparse vectors.
func sparsevecL2Distance(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(sparseVector), args[1].(sparseVector)
	if err := checkSparseDims(a, b); err != nil {
		return nil, err
	}
	return math.Sqrt(float64(sparseL2Squared(a, b))), nil
}

// sparsevecL2SquaredDistance implements sparsevec_l2_squared_distance.
func sparsevecL2SquaredDistance(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(sparseVector), args[1].(sparseVector)
	if err := checkSparseDims(a, b); err != nil {
		return nil, err
	}
	return float64(sparseL2Squared(a, b)), nil
}

// sparsevecInnerProduct implements inner_product for sparse vectors.
func sparsevecInnerProduct(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(sparseVector), args[1].(sparseVector)
	if err := checkSparseDims(a, b); err != nil {
		return nil, err
	}
	return float64(sparseInnerProduct(a, b)), nil
}

// sparsevecNegativeInnerProduct implements sparsevec_negative_inner_product.
func sparsevecNegativeInnerProduct(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(sparseVector), args[1].(sparseVector)
	if err := checkSparseDims(a, b); err != nil {
		return nil, err
	}
	return -float64(sparseInnerProduct(a, b)), nil
}

// sparsevecCosineDistance implements cosine_distance for sparse vectors, with the same mixed-precision arithmetic as
// the dense version.
func sparsevecCosineDistance(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(sparseVector), args[1].(sparseVector)
	if err := checkSparseDims(a, b); err != nil {
		return nil, err
	}
	var normA, normB float32
	for _, val := range a.Values {
		normA += val * val
	}
	for _, val := range b.Values {
		normB += val * val
	}
	similarity := float64(sparseInnerProduct(a, b)) / math.Sqrt(float64(normA)*float64(normB))
	if similarity > 1 {
		similarity = 1
	} else if similarity < -1 {
		similarity = -1
	}
	return 1 - similarity, nil
}

// sparsevecL1Distance implements l1_distance for sparse vectors, merging the two vectors in index order.
func sparsevecL1Distance(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(sparseVector), args[1].(sparseVector)
	if err := checkSparseDims(a, b); err != nil {
		return nil, err
	}
	distance := float32(0)
	i, j := 0, 0
	for i < len(a.Indices) && j < len(b.Indices) {
		switch {
		case a.Indices[i] == b.Indices[j]:
			distance += float32(math.Abs(float64(a.Values[i] - b.Values[j])))
			i++
			j++
		case a.Indices[i] < b.Indices[j]:
			distance += float32(math.Abs(float64(a.Values[i])))
			i++
		default:
			distance += float32(math.Abs(float64(b.Values[j])))
			j++
		}
	}
	for ; i < len(a.Indices); i++ {
		distance += float32(math.Abs(float64(a.Values[i])))
	}
	for ; j < len(b.Indices); j++ {
		distance += float32(math.Abs(float64(b.Values[j])))
	}
	return float64(distance), nil
}

// sparsevecL2Norm implements l2_norm for sparse vectors.
func sparsevecL2Norm(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	norm := float64(0)
	for _, val := range sv.Values {
		norm += float64(val) * float64(val)
	}
	return math.Sqrt(norm), nil
}

// sparsevecL2Normalize implements l2_normalize for sparse vectors, returning a zero vector unchanged.
func sparsevecL2Normalize(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	norm := float64(0)
	for _, val := range sv.Values {
		norm += float64(val) * float64(val)
	}
	norm = math.Sqrt(norm)
	if norm <= 0 {
		return sv, nil
	}
	result := sparseVector{Dim: sv.Dim}
	for i, val := range sv.Values {
		narrowed, err := checkFloat32Overflow(float32(float64(val) / norm))
		if err != nil {
			return nil, err
		}
		if narrowed != 0 {
			result.Indices = append(result.Indices, sv.Indices[i])
			result.Values = append(result.Values, narrowed)
		}
	}
	return result, nil
}

// sparseCompareValues compares two sparse vectors with dense semantics, treating missing indices as zeros and breaking
// ties on the dimension counts.
func sparseCompareValues(lhs any, rhs any) int32 {
	a, b := lhs.(sparseVector), rhs.(sparseVector)
	i, j := 0, 0
	for i < len(a.Indices) || j < len(b.Indices) {
		aVal, bVal := float32(0), float32(0)
		switch {
		case j >= len(b.Indices) || (i < len(a.Indices) && a.Indices[i] < b.Indices[j]):
			aVal = a.Values[i]
			i++
		case i >= len(a.Indices) || b.Indices[j] < a.Indices[i]:
			bVal = b.Values[j]
			j++
		default:
			aVal, bVal = a.Values[i], b.Values[j]
			i++
			j++
		}
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
	}
	if a.Dim < b.Dim {
		return -1
	} else if a.Dim > b.Dim {
		return 1
	}
	return 0
}

// sparsevecCastWithTypmod implements the identity cast function sparsevec(sparsevec, integer, boolean).
func sparsevecCastWithTypmod(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	if err := checkExpectedDims(args[1].(int32), int(sv.Dim)); err != nil {
		return nil, err
	}
	return sv, nil
}

// vectorToSparsevec implements vector_to_sparsevec.
func vectorToSparsevec(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	if err := checkExpectedDims(args[1].(int32), len(vals)); err != nil {
		return nil, err
	}
	return sparseFromDense(vals), nil
}

// sparsevecToVector implements sparsevec_to_vector.
func sparsevecToVector(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	if sv.Dim > maxDenseDims {
		return nil, errTooManyDims("vector", maxDenseDims)
	}
	if err := checkExpectedDims(args[1].(int32), int(sv.Dim)); err != nil {
		return nil, err
	}
	return sparseToDense(sv), nil
}

// halfvecToSparsevec implements halfvec_to_sparsevec, whose values are already exact float32s.
func halfvecToSparsevec(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	if err := checkExpectedDims(args[1].(int32), len(vals)); err != nil {
		return nil, err
	}
	return sparseFromDense(vals), nil
}

// sparsevecToHalfvec implements sparsevec_to_halfvec, which rounds every element to half precision.
func sparsevecToHalfvec(ctx *sql.Context, args ...any) (any, error) {
	sv := args[0].(sparseVector)
	if sv.Dim > maxDenseDims {
		return nil, errTooManyDims("halfvec", maxDenseDims)
	}
	if err := checkExpectedDims(args[1].(int32), int(sv.Dim)); err != nil {
		return nil, err
	}
	vals := sparseToDense(sv)
	for i, val := range vals {
		rounded, err := halfQuantize(val)
		if err != nil {
			return nil, err
		}
		vals[i] = rounded
	}
	return vals, nil
}

// arrayToSparsevec implements the array conversion overloads for the sparsevec type.
func arrayToSparsevec(ctx *sql.Context, args ...any) (any, error) {
	arr := args[0].([]any)
	if len(arr) == 0 {
		return nil, errAtLeastOneDim("sparsevec")
	}
	sv := sparseVector{Dim: int32(len(arr))}
	for i, elem := range arr {
		if elem == nil {
			return nil, errors.New("array must not contain nulls")
		}
		val, err := float32FromElement(elem)
		if err != nil {
			return nil, err
		}
		if err = checkElement("sparsevec", val); err != nil {
			return nil, err
		}
		if val != 0 {
			sv.Indices = append(sv.Indices, int32(i))
			sv.Values = append(sv.Values, val)
		}
	}
	if len(sv.Indices) > maxSparseNonZero {
		return nil, errors.Errorf("sparsevec cannot have more than %d non-zero elements", maxSparseNonZero)
	}
	if err := checkExpectedDims(args[1].(int32), int(sv.Dim)); err != nil {
		return nil, err
	}
	return sv, nil
}
