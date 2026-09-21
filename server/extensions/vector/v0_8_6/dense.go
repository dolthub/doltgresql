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
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// maxDenseDims is the dimension limit that pgvector enforces for the vector and halfvec types.
const maxDenseDims = 16000

// errMalformedLiteral reports a syntax error in a vector literal.
func errMalformedLiteral(typeName string, input string) error {
	return errors.Errorf(`invalid input syntax for type %s: "%s"`, typeName, input)
}

// errAtLeastOneDim reports a vector with no dimensions.
func errAtLeastOneDim(typeName string) error {
	return errors.Errorf("%s must have at least 1 dimension", typeName)
}

// errTooManyDims reports a vector that exceeds the named type's dimension limit.
func errTooManyDims(typeName string, maxDims int32) error {
	return errors.Errorf("%s cannot have more than %d dimensions", typeName, maxDims)
}

// errValueOutOfRange reports float overflow or underflow using PostgreSQL's wording.
func errValueOutOfRange(kind string) error {
	return errors.Errorf("value out of range: %s", kind)
}

// checkDims returns an error unless both vectors have the same number of dimensions.
func checkDims(typeName string, a []float32, b []float32) error {
	if len(a) != len(b) {
		return errors.Errorf("different %s dimensions %d and %d", typeName, len(a), len(b))
	}
	return nil
}

// checkDenseDim returns an error unless the dimension count is valid for the named dense type.
func checkDenseDim(typeName string, dim int) error {
	if dim < 1 {
		return errAtLeastOneDim(typeName)
	}
	if dim > maxDenseDims {
		return errTooManyDims(typeName, maxDenseDims)
	}
	return nil
}

// checkStateArray validates an aggregate [count, sums...] state array on behalf of the named caller, returning its
// values.
func checkStateArray(caller string, arr []any) ([]float64, error) {
	if len(arr) < 1 {
		return nil, errors.Errorf("%s: expected state array", caller)
	}
	vals := make([]float64, len(arr))
	for i, elem := range arr {
		val, ok := elem.(float64)
		if !ok {
			return nil, errors.Errorf("%s: expected state array", caller)
		}
		vals[i] = val
	}
	return vals, nil
}

// checkExpectedDims returns an error unless the dimension count satisfies the given type modifier.
func checkExpectedDims(typmod int32, dims int) error {
	if typmod != -1 && typmod != int32(dims) {
		return errors.Errorf("expected %d dimensions, not %d", typmod, dims)
	}
	return nil
}

// checkElement rejects the NaN and infinity element values that pgvector disallows.
func checkElement(typeName string, val float32) error {
	if math.IsNaN(float64(val)) {
		return errors.Errorf("NaN not allowed in %s", typeName)
	}
	if math.IsInf(float64(val), 0) {
		return errors.Errorf("infinite value not allowed in %s", typeName)
	}
	return nil
}

// checkFloat32Overflow returns an error when a float32 operation produced an infinity.
func checkFloat32Overflow(val float32) (float32, error) {
	if math.IsInf(float64(val), 0) {
		return 0, errValueOutOfRange("overflow")
	}
	return val, nil
}

// splitDenseLiteral splits a "[x,y,...]" literal into its trimmed element tokens.
func splitDenseLiteral(typeName string, input string) ([]string, error) {
	trimmed := strings.TrimSpace(input)
	end := strings.IndexByte(trimmed, ']')
	if !strings.HasPrefix(trimmed, "[") || end == -1 || len(strings.TrimSpace(trimmed[end+1:])) != 0 {
		return nil, errMalformedLiteral(typeName, input)
	}
	inner := strings.TrimSpace(trimmed[1:end])
	if len(inner) == 0 {
		return nil, errAtLeastOneDim(typeName)
	}
	tokens := strings.Split(inner, ",")
	if len(tokens) > maxDenseDims {
		return nil, errTooManyDims(typeName, maxDenseDims)
	}
	for i, token := range tokens {
		token = strings.TrimSpace(token)
		if len(token) == 0 {
			return nil, errMalformedLiteral(typeName, input)
		}
		tokens[i] = token
	}
	return tokens, nil
}

// parseElement parses a single element token as a float32, mirroring pgvector's strtof handling: overflow is an error
// while underflow silently loses precision.
func parseElement(typeName string, input string, token string) (float32, error) {
	fVal, err := strconv.ParseFloat(token, 32)
	if err != nil {
		if numErr, ok := err.(*strconv.NumError); !ok || numErr.Err != strconv.ErrRange {
			return 0, errMalformedLiteral(typeName, input)
		}
		if math.IsInf(fVal, 0) {
			return 0, errors.Errorf(`"%s" is out of range for type %s`, token, typeName)
		}
	}
	return float32(fVal), nil
}

// float32FromElement converts a supported array element to float32.
func float32FromElement(elem any) (float32, error) {
	switch val := elem.(type) {
	case int32:
		return float32(val), nil
	case float32:
		return val, nil
	case float64:
		return float32(val), nil
	case *apd.Decimal:
		fVal, err := val.Float64()
		if err != nil {
			return 0, err
		}
		return float32(fVal), nil
	default:
		return 0, errors.Errorf("unexpected array element type %T", elem)
	}
}

// typmodInValue parses the single dimensions modifier that every pgvector type accepts.
func typmodInValue(typeName string, maxDims int32, modifiers []any) (int32, error) {
	if len(modifiers) != 1 {
		return 0, errors.New("invalid type modifier")
	}
	str := modifiers[0].(string)
	dims, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
			return 0, errors.Errorf(`value "%s" is out of range for type integer`, str)
		}
		return 0, errors.Errorf(`invalid input syntax for type integer: "%s"`, str)
	}
	if dims < 1 {
		return 0, errors.Errorf("dimensions for type %s must be at least 1", typeName)
	}
	if dims > int64(maxDims) {
		return 0, errors.Errorf("dimensions for type %s cannot exceed %d", typeName, maxDims)
	}
	return int32(dims), nil
}

// formatElement renders a single element the way PostgreSQL renders a float4.
func formatElement(val float32) string {
	return strconv.FormatFloat(float64(val), 'g', -1, 32)
}

// formatDense renders a dense vector as its "[x,y,...]" text form.
func formatDense(vals []float32) string {
	sb := strings.Builder{}
	sb.WriteByte('[')
	for i, val := range vals {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(formatElement(val))
	}
	sb.WriteByte(']')
	return sb.String()
}

// denseL2Squared returns the squared Euclidean distance, accumulating in float32 like pgvector.
func denseL2Squared(a []float32, b []float32) float32 {
	distance := float32(0)
	for i := range a {
		diff := a[i] - b[i]
		distance += diff * diff
	}
	return distance
}

// denseInnerProduct returns the inner product, accumulating in float32 like pgvector.
func denseInnerProduct(a []float32, b []float32) float32 {
	product := float32(0)
	for i := range a {
		product += a[i] * b[i]
	}
	return product
}

// denseCosineDistance returns the cosine distance, mirroring pgvector's mixed-precision arithmetic: float32
// accumulation, float64 division, and a clamp of the similarity into [-1, 1] that lets NaN through.
func denseCosineDistance(a []float32, b []float32) float64 {
	var product, normA, normB float32
	for i := range a {
		product += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	similarity := float64(product) / math.Sqrt(float64(normA)*float64(normB))
	if similarity > 1 {
		similarity = 1
	} else if similarity < -1 {
		similarity = -1
	}
	return 1 - similarity
}

// denseL1Distance returns the taxicab distance, accumulating in float32 like pgvector.
func denseL1Distance(a []float32, b []float32) float32 {
	distance := float32(0)
	for i := range a {
		distance += float32(math.Abs(float64(a[i] - b[i])))
	}
	return distance
}

// denseNorm returns the Euclidean norm, accumulating in float64 like pgvector.
func denseNorm(vals []float32) float64 {
	norm := float64(0)
	for _, val := range vals {
		norm += float64(val) * float64(val)
	}
	return math.Sqrt(norm)
}

// denseCompare compares two vectors lexicographically, ordering a shorter vector before a longer one that it prefixes.
func denseCompare(a []float32, b []float32) int32 {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}
	return 0
}

// denseCompareValues adapts denseCompare to the signature that comparisonRoutines expects.
func denseCompareValues(a any, b any) int32 {
	return denseCompare(a.([]float32), b.([]float32))
}

// denseL2DistanceImpl returns an implementation of l2_distance for the named dense type.
func denseL2DistanceImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		return math.Sqrt(float64(denseL2Squared(a, b))), nil
	}
}

// denseL2SquaredDistanceImpl returns an implementation of the squared distance support function for the named type.
func denseL2SquaredDistanceImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		return float64(denseL2Squared(a, b)), nil
	}
}

// denseInnerProductImpl returns an implementation of inner_product for the named dense type.
func denseInnerProductImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		return float64(denseInnerProduct(a, b)), nil
	}
}

// denseNegativeInnerProductImpl returns an implementation of the negative inner product for the named dense type.
func denseNegativeInnerProductImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		return -float64(denseInnerProduct(a, b)), nil
	}
}

// denseCosineDistanceImpl returns an implementation of cosine_distance for the named dense type.
func denseCosineDistanceImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		return denseCosineDistance(a, b), nil
	}
}

// denseSphericalDistanceImpl returns an implementation of the spherical distance support function for the named type.
func denseSphericalDistanceImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		distance := float64(denseInnerProduct(a, b))
		if distance > 1 {
			distance = 1
		} else if distance < -1 {
			distance = -1
		}
		return math.Acos(distance) / math.Pi, nil
	}
}

// denseL1DistanceImpl returns an implementation of l1_distance for the named dense type.
func denseL1DistanceImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		return float64(denseL1Distance(a, b)), nil
	}
}

// denseDims implements vector_dims for both dense types.
func denseDims(ctx *sql.Context, args ...any) (any, error) {
	return int32(len(args[0].([]float32))), nil
}

// denseNormValue implements vector_norm and l2_norm for the dense types.
func denseNormValue(ctx *sql.Context, args ...any) (any, error) {
	return denseNorm(args[0].([]float32)), nil
}

// denseL2NormalizeImpl returns an implementation of l2_normalize for a dense type, narrowing each scaled element
// through the given function. A zero vector is returned unchanged.
func denseL2NormalizeImpl(narrow func(float32) (float32, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		vals := args[0].([]float32)
		norm := denseNorm(vals)
		if norm <= 0 {
			return vals, nil
		}
		result := make([]float32, len(vals))
		for i, val := range vals {
			narrowed, err := narrow(float32(float64(val) / norm))
			if err != nil {
				return nil, err
			}
			result[i] = narrowed
		}
		return result, nil
	}
}

// denseBinaryQuantize implements binary_quantize for both dense types, setting a bit for every positive element.
func denseBinaryQuantize(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	bits := make([]byte, len(vals))
	for i, val := range vals {
		if val > 0 {
			bits[i] = '1'
		} else {
			bits[i] = '0'
		}
	}
	return string(bits), nil
}

// denseSubvectorImpl returns an implementation of subvector for the named dense type, extracting count dimensions
// from the 1-based start index with pgvector's clamping rules.
func denseSubvectorImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		vals := args[0].([]float32)
		start, count := args[1].(int32), args[2].(int32)
		if count < 1 {
			return nil, errAtLeastOneDim(typeName)
		}
		dim := int32(len(vals))
		var end int32
		if start > dim-count+1 {
			end = dim + 1
		} else {
			end = start + count
		}
		if start < 1 {
			start = 1
		} else if start > dim {
			return nil, errAtLeastOneDim(typeName)
		}
		if err := checkDenseDim(typeName, int(end-start)); err != nil {
			return nil, err
		}
		return vals[start-1 : end-1], nil
	}
}

// denseArithmeticImpl returns a pairwise arithmetic implementation for the named dense type, narrowing each result
// through the given function. checkUnderflow additionally rejects results that vanished, which multiplication requires.
func denseArithmeticImpl(typeName string, op func(x float32, y float32) float32, narrow func(float32) (float32, error), checkUnderflow bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if err := checkDims(typeName, a, b); err != nil {
			return nil, err
		}
		result := make([]float32, len(a))
		for i := range a {
			val, err := narrow(op(a[i], b[i]))
			if err != nil {
				return nil, err
			}
			if checkUnderflow && val == 0 && a[i] != 0 && b[i] != 0 {
				return nil, errValueOutOfRange("underflow")
			}
			result[i] = val
		}
		return result, nil
	}
}

// denseConcatImpl returns an implementation of the concatenation function for the named dense type.
func denseConcatImpl(typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].([]float32), args[1].([]float32)
		if len(a)+len(b) > maxDenseDims {
			return nil, errTooManyDims(typeName, maxDenseDims)
		}
		return append(append(make([]float32, 0, len(a)+len(b)), a...), b...), nil
	}
}

// denseAccumImpl returns the avg transition function for the named caller, adding a vector into the
// [count, sums...] state. A single-element state is empty and adopts the vector's dimensions.
func denseAccumImpl(caller string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		state, err := checkStateArray(caller, args[0].([]any))
		if err != nil {
			return nil, err
		}
		vals := args[1].([]float32)
		if len(state) > 1 && len(state)-1 != len(vals) {
			return nil, errors.Errorf("expected %d dimensions, not %d", len(state)-1, len(vals))
		}
		newState := make([]any, len(vals)+1)
		newState[0] = state[0] + 1
		if len(state) == 1 {
			for i, val := range vals {
				newState[i+1] = float64(val)
			}
			return newState, nil
		}
		for i, val := range vals {
			sum := state[i+1] + float64(val)
			if math.IsInf(sum, 0) {
				return nil, errValueOutOfRange("overflow")
			}
			newState[i+1] = sum
		}
		return newState, nil
	}
}

// denseCombineImpl returns the avg combine function for the named caller and dense type, merging two
// [count, sums...] states. An empty state takes the other state's value.
func denseCombineImpl(caller string, typeName string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, err := checkStateArray(caller, args[0].([]any))
		if err != nil {
			return nil, err
		}
		b, err := checkStateArray(caller, args[1].([]any))
		if err != nil {
			return nil, err
		}
		if a[0] == 0 {
			a, b = b, a
		}
		if err = checkDenseDim(typeName, len(a)-1); err != nil {
			return nil, err
		}
		result := make([]any, len(a))
		if b[0] == 0 {
			for i, val := range a {
				result[i] = val
			}
			return result, nil
		}
		if len(a) != len(b) {
			return nil, errors.Errorf("expected %d dimensions, not %d", len(a)-1, len(b)-1)
		}
		result[0] = a[0] + b[0]
		for i := 1; i < len(a); i++ {
			sum := a[i] + b[i]
			if math.IsInf(sum, 0) {
				return nil, errValueOutOfRange("overflow")
			}
			result[i] = sum
		}
		return result, nil
	}
}

// denseAvgImpl returns the avg final function for the named caller and dense type, narrowing each averaged element
// through the given function and returning NULL for an empty group.
func denseAvgImpl(caller string, typeName string, narrow func(float32) (float32, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		state, err := checkStateArray(caller, args[0].([]any))
		if err != nil {
			return nil, err
		}
		n := state[0]
		if n == 0 {
			return nil, nil
		}
		if err = checkDenseDim(typeName, len(state)-1); err != nil {
			return nil, err
		}
		result := make([]float32, len(state)-1)
		for i := range result {
			val, err := narrow(float32(state[i+1] / n))
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	}
}

// denseCastWithTypmod implements the identity cast functions vector(vector, integer, boolean) and
// halfvec(halfvec, integer, boolean), which only enforce the target's type modifier.
func denseCastWithTypmod(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	if err := checkExpectedDims(args[1].(int32), len(vals)); err != nil {
		return nil, err
	}
	return vals, nil
}

// arrayToDenseImpl returns an implementation of the array conversion overloads for the named dense type, narrowing
// each element through the given function when one is provided.
func arrayToDenseImpl(typeName string, narrow func(float32) (float32, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		arr := args[0].([]any)
		if len(arr) == 0 {
			return nil, errAtLeastOneDim(typeName)
		}
		if len(arr) > maxDenseDims {
			return nil, errTooManyDims(typeName, maxDenseDims)
		}
		result := make([]float32, len(arr))
		for i, elem := range arr {
			if elem == nil {
				return nil, errors.New("array must not contain nulls")
			}
			val, err := float32FromElement(elem)
			if err != nil {
				return nil, err
			}
			if err = checkElement(typeName, val); err != nil {
				return nil, err
			}
			if narrow != nil {
				if val, err = narrow(val); err != nil {
					return nil, err
				}
			}
			result[i] = val
		}
		if err := checkExpectedDims(args[1].(int32), len(result)); err != nil {
			return nil, err
		}
		return result, nil
	}
}

// denseToFloat4 implements the casts from the dense types to real[].
func denseToFloat4(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	result := make([]any, len(vals))
	for i, val := range vals {
		result[i] = val
	}
	return result, nil
}
