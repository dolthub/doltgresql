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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function/vector"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Extension returns the definition of the emulated extension. The access methods' handler and internal support
// functions are not declared, as both emulated methods map to Dolt's native vector index.
func Extension() *extdef.Extension {
	return &extdef.Extension{
		Name: "vector",
		Control: extdef.Control{
			DefaultVersion: "0.8.6",
			Comment:        "vector data type and ivfflat and hnsw access methods",
			Superuser:      true,
			Relocatable:    true,
		},
		Types: []extdef.Type{
			pgvectorType("vector", typeCodec(vectorSerialize, vectorDeserialize, denseCompareValues, true)),
			pgvectorType("halfvec", typeCodec(halfvecSerialize, halfvecDeserialize, denseCompareValues, true)),
			pgvectorType("sparsevec", typeCodec(sparsevecSerialize, sparsevecDeserialize, sparseCompareValues, false)),
		},
		Routines:        routines(),
		Operators:       operators(),
		Casts:           castList(),
		Aggregates:      aggregates(),
		OperatorClasses: operatorClasses(),
		AccessMethods:   accessMethods(),
	}
}

// pgvectorType declares one of the extension's types, whose support functions all follow the same naming scheme.
func pgvectorType(name string, codec *pgtypes.TypeCodec) extdef.Type {
	def := pgtypes.NewBaseTypeDefinition()
	def.Storage = pgtypes.TypeStorage_External
	return extdef.Type{
		Name:       name,
		Definition: def,
		Input:      name + "_in",
		Output:     name + "_out",
		Receive:    name + "_recv",
		Send:       name + "_send",
		ModIn:      name + "_typmod_in",
		Compare:    name + "_cmp",
		Codec:      codec,
	}
}

// params builds an unnamed parameter list from the given type names.
func params(typeNames ...string) []extdef.Parameter {
	ps := make([]extdef.Parameter, len(typeNames))
	for i, typeName := range typeNames {
		ps[i] = extdef.Parameter{Type: typeName}
	}
	return ps
}

// ioRoutines returns the input, output, type modifier, receive, and send routines for one of the extension's types.
func ioRoutines(typeName string, in extdef.Function, out extdef.Function, typmodIn extdef.Function, recv extdef.Function, send extdef.Function) []extdef.Routine {
	return []extdef.Routine{
		{Name: typeName + "_in", Symbol: typeName + "_in", Parameters: params("cstring", "oid", "int4"), Returns: typeName, Strict: true, Impl: in},
		{Name: typeName + "_out", Symbol: typeName + "_out", Parameters: params(typeName), Returns: "cstring", Strict: true, Impl: out},
		{Name: typeName + "_typmod_in", Symbol: typeName + "_typmod_in", Parameters: params("_cstring"), Returns: "int4", Strict: true, Impl: typmodIn},
		{Name: typeName + "_recv", Symbol: typeName + "_recv", Parameters: params("internal", "oid", "int4"), Returns: typeName, Strict: true, Impl: recv},
		{Name: typeName + "_send", Symbol: typeName + "_send", Parameters: params(typeName), Returns: "bytea", Strict: true, Impl: send},
	}
}

// comparisonRoutines returns the six comparison functions and the cmp function for the named type, deriving each from
// the given comparison.
func comparisonRoutines(typeName string, cmp func(a any, b any) int32) []extdef.Routine {
	tests := []struct {
		suffix  string
		matches func(int32) bool
	}{
		{"_lt", func(c int32) bool { return c < 0 }},
		{"_le", func(c int32) bool { return c <= 0 }},
		{"_eq", func(c int32) bool { return c == 0 }},
		{"_ne", func(c int32) bool { return c != 0 }},
		{"_ge", func(c int32) bool { return c >= 0 }},
		{"_gt", func(c int32) bool { return c > 0 }},
	}
	routines := make([]extdef.Routine, 0, len(tests)+1)
	for _, test := range tests {
		matches := test.matches
		name := typeName + test.suffix
		routines = append(routines, extdef.Routine{
			Name: name, Symbol: name, Parameters: params(typeName, typeName), Returns: "bool", Strict: true,
			Impl: func(ctx *sql.Context, args ...any) (any, error) {
				return matches(cmp(args[0], args[1])), nil
			},
		})
	}
	return append(routines, extdef.Routine{
		Name: typeName + "_cmp", Symbol: typeName + "_cmp", Parameters: params(typeName, typeName), Returns: "int4", Strict: true,
		Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return cmp(args[0], args[1]), nil
		},
	})
}

// arrayRoutines returns the four array conversion overloads for the named type. The overloads share the C symbol
// "array_to_<type>" in the real extension, so each takes its source type's name as a suffix to stay unique.
func arrayRoutines(typeName string, impl extdef.Function) []extdef.Routine {
	sources := []string{"_int4", "_float4", "_float8", "_numeric"}
	routines := make([]extdef.Routine, len(sources))
	for i, source := range sources {
		routines[i] = extdef.Routine{
			Name:       "array_to_" + typeName,
			Symbol:     "array_to_" + typeName + source,
			Parameters: params(source, "int4", "bool"),
			Returns:    typeName,
			Strict:     true,
			Impl:       impl,
		}
	}
	return routines
}

// routines returns every routine that the extension declares.
func routines() []extdef.Routine {
	addOp := func(x float32, y float32) float32 { return x + y }
	subOp := func(x float32, y float32) float32 { return x - y }
	mulOp := func(x float32, y float32) float32 { return x * y }
	routines := ioRoutines("vector", vectorIn, denseOut, vectorTypmodIn, vectorRecv, vectorSend)
	routines = append(routines,
		extdef.Routine{Name: "binary_quantize", Symbol: "binary_quantize", Parameters: params("vector"), Returns: "bit", Strict: true, Impl: denseBinaryQuantize},
		extdef.Routine{Name: "cosine_distance", Symbol: "cosine_distance", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseCosineDistanceImpl("vector"), DistanceType: vector.DistanceCosine{}},
		extdef.Routine{Name: "inner_product", Symbol: "inner_product", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseInnerProductImpl("vector")},
		extdef.Routine{Name: "l1_distance", Symbol: "l1_distance", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseL1DistanceImpl("vector"), DistanceType: vector.DistanceL1{}},
		extdef.Routine{Name: "l2_distance", Symbol: "l2_distance", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseL2DistanceImpl("vector"), DistanceType: vector.DistanceEuclidean{}},
		extdef.Routine{Name: "l2_normalize", Symbol: "l2_normalize", Parameters: params("vector"), Returns: "vector", Strict: true, Impl: denseL2NormalizeImpl(checkFloat32Overflow)},
		extdef.Routine{Name: "subvector", Symbol: "subvector", Parameters: params("vector", "int4", "int4"), Returns: "vector", Strict: true, Impl: denseSubvectorImpl("vector")},
		extdef.Routine{Name: "vector_add", Symbol: "vector_add", Parameters: params("vector", "vector"), Returns: "vector", Strict: true, Impl: denseArithmeticImpl("vector", addOp, checkFloat32Overflow, false)},
		extdef.Routine{Name: "vector_concat", Symbol: "vector_concat", Parameters: params("vector", "vector"), Returns: "vector", Strict: true, Impl: denseConcatImpl("vector")},
		extdef.Routine{Name: "vector_dims", Symbol: "vector_dims", Parameters: params("vector"), Returns: "int4", Strict: true, Impl: denseDims},
		extdef.Routine{Name: "vector_mul", Symbol: "vector_mul", Parameters: params("vector", "vector"), Returns: "vector", Strict: true, Impl: denseArithmeticImpl("vector", mulOp, checkFloat32Overflow, true)},
		extdef.Routine{Name: "vector_norm", Symbol: "vector_norm", Parameters: params("vector"), Returns: "float8", Strict: true, Impl: denseNormValue},
		extdef.Routine{Name: "vector_sub", Symbol: "vector_sub", Parameters: params("vector", "vector"), Returns: "vector", Strict: true, Impl: denseArithmeticImpl("vector", subOp, checkFloat32Overflow, false)},
	)
	routines = append(routines, comparisonRoutines("vector", denseCompareValues)...)
	routines = append(routines,
		extdef.Routine{Name: "vector", Symbol: "vector", Parameters: params("vector", "int4", "bool"), Returns: "vector", Strict: true, Impl: denseCastWithTypmod},
		extdef.Routine{Name: "vector_accum", Symbol: "vector_accum", Parameters: params("_float8", "vector"), Returns: "_float8", Strict: true, Impl: denseAccumImpl("vector_accum")},
		extdef.Routine{Name: "vector_avg", Symbol: "vector_avg", Parameters: params("_float8"), Returns: "vector", Strict: true, Impl: denseAvgImpl("vector_avg", "vector", checkFloat32Overflow)},
		extdef.Routine{Name: "vector_combine", Symbol: "vector_combine", Parameters: params("_float8", "_float8"), Returns: "_float8", Strict: true, Impl: denseCombineImpl("vector_combine", "vector")},
		extdef.Routine{Name: "vector_l2_squared_distance", Symbol: "vector_l2_squared_distance", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseL2SquaredDistanceImpl("vector")},
		extdef.Routine{Name: "vector_negative_inner_product", Symbol: "vector_negative_inner_product", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseNegativeInnerProductImpl("vector"), DistanceType: vector.DistanceInnerProduct{}},
		extdef.Routine{Name: "vector_spherical_distance", Symbol: "vector_spherical_distance", Parameters: params("vector", "vector"), Returns: "float8", Strict: true, Impl: denseSphericalDistanceImpl("vector")},
		extdef.Routine{Name: "vector_to_float4", Symbol: "vector_to_float4", Parameters: params("vector", "int4", "bool"), Returns: "_float4", Strict: true, Impl: denseToFloat4},
	)
	routines = append(routines, arrayRoutines("vector", arrayToDenseImpl("vector", nil))...)

	routines = append(routines, ioRoutines("halfvec", halfvecIn, denseOut, halfvecTypmodIn, halfvecRecv, halfvecSend)...)
	routines = append(routines,
		extdef.Routine{Name: "binary_quantize", Symbol: "halfvec_binary_quantize", Parameters: params("halfvec"), Returns: "bit", Strict: true, Impl: denseBinaryQuantize},
		extdef.Routine{Name: "cosine_distance", Symbol: "halfvec_cosine_distance", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseCosineDistanceImpl("halfvec"), DistanceType: vector.DistanceCosine{}},
		extdef.Routine{Name: "halfvec_add", Symbol: "halfvec_add", Parameters: params("halfvec", "halfvec"), Returns: "halfvec", Strict: true, Impl: denseArithmeticImpl("halfvec", addOp, halfQuantize, false)},
		extdef.Routine{Name: "halfvec_concat", Symbol: "halfvec_concat", Parameters: params("halfvec", "halfvec"), Returns: "halfvec", Strict: true, Impl: denseConcatImpl("halfvec")},
		extdef.Routine{Name: "halfvec_mul", Symbol: "halfvec_mul", Parameters: params("halfvec", "halfvec"), Returns: "halfvec", Strict: true, Impl: denseArithmeticImpl("halfvec", mulOp, halfQuantize, true)},
		extdef.Routine{Name: "halfvec_sub", Symbol: "halfvec_sub", Parameters: params("halfvec", "halfvec"), Returns: "halfvec", Strict: true, Impl: denseArithmeticImpl("halfvec", subOp, halfQuantize, false)},
		extdef.Routine{Name: "inner_product", Symbol: "halfvec_inner_product", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseInnerProductImpl("halfvec")},
		extdef.Routine{Name: "l1_distance", Symbol: "halfvec_l1_distance", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseL1DistanceImpl("halfvec"), DistanceType: vector.DistanceL1{}},
		extdef.Routine{Name: "l2_distance", Symbol: "halfvec_l2_distance", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseL2DistanceImpl("halfvec"), DistanceType: vector.DistanceEuclidean{}},
		extdef.Routine{Name: "l2_norm", Symbol: "halfvec_l2_norm", Parameters: params("halfvec"), Returns: "float8", Strict: true, Impl: denseNormValue},
		extdef.Routine{Name: "l2_normalize", Symbol: "halfvec_l2_normalize", Parameters: params("halfvec"), Returns: "halfvec", Strict: true, Impl: denseL2NormalizeImpl(halfQuantize)},
		extdef.Routine{Name: "subvector", Symbol: "halfvec_subvector", Parameters: params("halfvec", "int4", "int4"), Returns: "halfvec", Strict: true, Impl: denseSubvectorImpl("halfvec")},
		extdef.Routine{Name: "vector_dims", Symbol: "halfvec_vector_dims", Parameters: params("halfvec"), Returns: "int4", Strict: true, Impl: denseDims},
	)
	routines = append(routines, comparisonRoutines("halfvec", denseCompareValues)...)
	routines = append(routines,
		extdef.Routine{Name: "halfvec", Symbol: "halfvec", Parameters: params("halfvec", "int4", "bool"), Returns: "halfvec", Strict: true, Impl: denseCastWithTypmod},
		extdef.Routine{Name: "halfvec_accum", Symbol: "halfvec_accum", Parameters: params("_float8", "halfvec"), Returns: "_float8", Strict: true, Impl: denseAccumImpl("halfvec_accum")},
		extdef.Routine{Name: "halfvec_avg", Symbol: "halfvec_avg", Parameters: params("_float8"), Returns: "halfvec", Strict: true, Impl: denseAvgImpl("halfvec_avg", "halfvec", halfQuantize)},
		extdef.Routine{Name: "halfvec_combine", Symbol: "halfvec_combine", Parameters: params("_float8", "_float8"), Returns: "_float8", Strict: true, Impl: denseCombineImpl("halfvec_combine", "halfvec")},
		extdef.Routine{Name: "halfvec_l2_squared_distance", Symbol: "halfvec_l2_squared_distance", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseL2SquaredDistanceImpl("halfvec")},
		extdef.Routine{Name: "halfvec_negative_inner_product", Symbol: "halfvec_negative_inner_product", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseNegativeInnerProductImpl("halfvec"), DistanceType: vector.DistanceInnerProduct{}},
		extdef.Routine{Name: "halfvec_spherical_distance", Symbol: "halfvec_spherical_distance", Parameters: params("halfvec", "halfvec"), Returns: "float8", Strict: true, Impl: denseSphericalDistanceImpl("halfvec")},
		extdef.Routine{Name: "halfvec_to_float4", Symbol: "halfvec_to_float4", Parameters: params("halfvec", "int4", "bool"), Returns: "_float4", Strict: true, Impl: denseToFloat4},
		extdef.Routine{Name: "halfvec_to_vector", Symbol: "halfvec_to_vector", Parameters: params("halfvec", "int4", "bool"), Returns: "vector", Strict: true, Impl: halfvecToVector},
		extdef.Routine{Name: "vector_to_halfvec", Symbol: "vector_to_halfvec", Parameters: params("vector", "int4", "bool"), Returns: "halfvec", Strict: true, Impl: vectorToHalfvec},
	)
	routines = append(routines, arrayRoutines("halfvec", arrayToDenseImpl("halfvec", halfQuantize))...)

	routines = append(routines,
		extdef.Routine{Name: "hamming_distance", Symbol: "hamming_distance", Parameters: params("bit", "bit"), Returns: "float8", Strict: true, Impl: hammingDistance},
		extdef.Routine{Name: "jaccard_distance", Symbol: "jaccard_distance", Parameters: params("bit", "bit"), Returns: "float8", Strict: true, Impl: jaccardDistance},
	)

	routines = append(routines, ioRoutines("sparsevec", sparsevecIn, sparsevecOut, sparsevecTypmodIn, sparsevecRecv, sparsevecSend)...)
	routines = append(routines,
		extdef.Routine{Name: "cosine_distance", Symbol: "sparsevec_cosine_distance", Parameters: params("sparsevec", "sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecCosineDistance},
		extdef.Routine{Name: "inner_product", Symbol: "sparsevec_inner_product", Parameters: params("sparsevec", "sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecInnerProduct},
		extdef.Routine{Name: "l1_distance", Symbol: "sparsevec_l1_distance", Parameters: params("sparsevec", "sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecL1Distance},
		extdef.Routine{Name: "l2_distance", Symbol: "sparsevec_l2_distance", Parameters: params("sparsevec", "sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecL2Distance},
		extdef.Routine{Name: "l2_norm", Symbol: "sparsevec_l2_norm", Parameters: params("sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecL2Norm},
		extdef.Routine{Name: "l2_normalize", Symbol: "sparsevec_l2_normalize", Parameters: params("sparsevec"), Returns: "sparsevec", Strict: true, Impl: sparsevecL2Normalize},
	)
	routines = append(routines, comparisonRoutines("sparsevec", sparseCompareValues)...)
	routines = append(routines,
		extdef.Routine{Name: "halfvec_to_sparsevec", Symbol: "halfvec_to_sparsevec", Parameters: params("halfvec", "int4", "bool"), Returns: "sparsevec", Strict: true, Impl: halfvecToSparsevec},
		extdef.Routine{Name: "sparsevec", Symbol: "sparsevec", Parameters: params("sparsevec", "int4", "bool"), Returns: "sparsevec", Strict: true, Impl: sparsevecCastWithTypmod},
		extdef.Routine{Name: "sparsevec_l2_squared_distance", Symbol: "sparsevec_l2_squared_distance", Parameters: params("sparsevec", "sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecL2SquaredDistance},
		extdef.Routine{Name: "sparsevec_negative_inner_product", Symbol: "sparsevec_negative_inner_product", Parameters: params("sparsevec", "sparsevec"), Returns: "float8", Strict: true, Impl: sparsevecNegativeInnerProduct},
		extdef.Routine{Name: "sparsevec_to_halfvec", Symbol: "sparsevec_to_halfvec", Parameters: params("sparsevec", "int4", "bool"), Returns: "halfvec", Strict: true, Impl: sparsevecToHalfvec},
		extdef.Routine{Name: "sparsevec_to_vector", Symbol: "sparsevec_to_vector", Parameters: params("sparsevec", "int4", "bool"), Returns: "vector", Strict: true, Impl: sparsevecToVector},
		extdef.Routine{Name: "vector_to_sparsevec", Symbol: "vector_to_sparsevec", Parameters: params("vector", "int4", "bool"), Returns: "sparsevec", Strict: true, Impl: vectorToSparsevec},
	)
	return append(routines, arrayRoutines("sparsevec", arrayToSparsevec)...)
}

// distanceOperators returns the four distance operators for the named type, referencing the given routine symbols.
func distanceOperators(typeName string, l2 string, negInnerProduct string, cosine string, l1 string) []extdef.Operator {
	return []extdef.Operator{
		{Symbol: "<->", Left: typeName, Right: typeName, Routine: l2, Commutator: "<->"},
		{Symbol: "<#>", Left: typeName, Right: typeName, Routine: negInnerProduct, Commutator: "<#>"},
		{Symbol: "<=>", Left: typeName, Right: typeName, Routine: cosine, Commutator: "<=>"},
		{Symbol: "<+>", Left: typeName, Right: typeName, Routine: l1, Commutator: "<+>"},
	}
}

// arithmeticOperators returns the arithmetic and concatenation operators for the named dense type.
func arithmeticOperators(typeName string) []extdef.Operator {
	return []extdef.Operator{
		{Symbol: "+", Left: typeName, Right: typeName, Routine: typeName + "_add", Commutator: "+"},
		{Symbol: "-", Left: typeName, Right: typeName, Routine: typeName + "_sub"},
		{Symbol: "*", Left: typeName, Right: typeName, Routine: typeName + "_mul", Commutator: "*"},
		{Symbol: "||", Left: typeName, Right: typeName, Routine: typeName + "_concat"},
	}
}

// comparisonOperators returns the six comparison operators for the named type.
func comparisonOperators(typeName string) []extdef.Operator {
	return []extdef.Operator{
		{Symbol: "<", Left: typeName, Right: typeName, Routine: typeName + "_lt", Commutator: ">", Negator: ">="},
		{Symbol: "<=", Left: typeName, Right: typeName, Routine: typeName + "_le", Commutator: ">=", Negator: ">"},
		{Symbol: "=", Left: typeName, Right: typeName, Routine: typeName + "_eq", Commutator: "=", Negator: "<>"},
		{Symbol: "<>", Left: typeName, Right: typeName, Routine: typeName + "_ne", Commutator: "<>", Negator: "="},
		{Symbol: ">=", Left: typeName, Right: typeName, Routine: typeName + "_ge", Commutator: "<=", Negator: "<"},
		{Symbol: ">", Left: typeName, Right: typeName, Routine: typeName + "_gt", Commutator: "<", Negator: "<="},
	}
}

// operators returns every operator that the extension declares.
func operators() []extdef.Operator {
	ops := distanceOperators("vector", "l2_distance", "vector_negative_inner_product", "cosine_distance", "l1_distance")
	ops = append(ops, arithmeticOperators("vector")...)
	ops = append(ops, comparisonOperators("vector")...)
	ops = append(ops, distanceOperators("halfvec", "halfvec_l2_distance", "halfvec_negative_inner_product", "halfvec_cosine_distance", "halfvec_l1_distance")...)
	ops = append(ops, arithmeticOperators("halfvec")...)
	ops = append(ops, comparisonOperators("halfvec")...)
	ops = append(ops,
		extdef.Operator{Symbol: "<~>", Left: "bit", Right: "bit", Routine: "hamming_distance", Commutator: "<~>"},
		extdef.Operator{Symbol: "<%>", Left: "bit", Right: "bit", Routine: "jaccard_distance", Commutator: "<%>"},
	)
	ops = append(ops, distanceOperators("sparsevec", "sparsevec_l2_distance", "sparsevec_negative_inner_product", "sparsevec_cosine_distance", "sparsevec_l1_distance")...)
	return append(ops, comparisonOperators("sparsevec")...)
}

// arrayCasts returns the assignment casts from the four supported array types to the named type.
func arrayCasts(typeName string) []extdef.Cast {
	sources := []string{"_int4", "_float4", "_float8", "_numeric"}
	list := make([]extdef.Cast, len(sources))
	for i, source := range sources {
		list[i] = extdef.Cast{
			Source:   source,
			Target:   typeName,
			Routine:  "array_to_" + typeName + source,
			CastType: casts.CastType_Assignment,
		}
	}
	return list
}

// castList returns every cast that the extension declares.
func castList() []extdef.Cast {
	list := []extdef.Cast{
		{Source: "halfvec", Target: "_float4", Routine: "halfvec_to_float4", CastType: casts.CastType_Assignment},
		{Source: "halfvec", Target: "halfvec", Routine: "halfvec", CastType: casts.CastType_Implicit},
		{Source: "halfvec", Target: "sparsevec", Routine: "halfvec_to_sparsevec", CastType: casts.CastType_Implicit},
		{Source: "halfvec", Target: "vector", Routine: "halfvec_to_vector", CastType: casts.CastType_Assignment},
		{Source: "sparsevec", Target: "halfvec", Routine: "sparsevec_to_halfvec", CastType: casts.CastType_Assignment},
		{Source: "sparsevec", Target: "sparsevec", Routine: "sparsevec", CastType: casts.CastType_Implicit},
		{Source: "sparsevec", Target: "vector", Routine: "sparsevec_to_vector", CastType: casts.CastType_Assignment},
		{Source: "vector", Target: "_float4", Routine: "vector_to_float4", CastType: casts.CastType_Implicit},
		{Source: "vector", Target: "halfvec", Routine: "vector_to_halfvec", CastType: casts.CastType_Implicit},
		{Source: "vector", Target: "sparsevec", Routine: "vector_to_sparsevec", CastType: casts.CastType_Implicit},
		{Source: "vector", Target: "vector", Routine: "vector", CastType: casts.CastType_Implicit},
	}
	list = append(list, arrayCasts("vector")...)
	list = append(list, arrayCasts("halfvec")...)
	return append(list, arrayCasts("sparsevec")...)
}

// operatorClasses returns the extension's operator classes. Postgres offers the dense L1 classes under hnsw only, but
// both emulated methods map to the same index, so every dense class is declared under both. The sparsevec and bit
// classes are declared for the catalog, but index support for them is not yet implemented.
func operatorClasses() []extdef.OperatorClass {
	both := []string{"hnsw", "ivfflat"}
	hnswOnly := []string{"hnsw"}
	metrics := []struct {
		suffix       string
		distanceType vector.DistanceType
	}{
		{"_l2_ops", vector.DistanceEuclidean{}},
		{"_ip_ops", vector.DistanceInnerProduct{}},
		{"_cosine_ops", vector.DistanceCosine{}},
		{"_l1_ops", vector.DistanceL1{}},
	}
	opclasses := make([]extdef.OperatorClass, 0, 14)
	for _, dense := range []struct {
		typeName      string
		maxDimensions int32
	}{{"vector", 2000}, {"halfvec", 4000}} {
		for _, metric := range metrics {
			opclass := extdef.OperatorClass{
				Name:          dense.typeName + metric.suffix,
				AccessMethods: both,
				Type:          dense.typeName,
				DistanceType:  metric.distanceType,
				MaxDimensions: dense.maxDimensions,
			}
			if opclass.Name == "vector_l2_ops" {
				opclass.DefaultFor = []string{"ivfflat"}
			}
			opclasses = append(opclasses, opclass)
		}
	}
	for _, metric := range metrics {
		opclasses = append(opclasses, extdef.OperatorClass{
			Name:          "sparsevec" + metric.suffix,
			AccessMethods: hnswOnly,
			Type:          "sparsevec",
		})
	}
	return append(opclasses,
		extdef.OperatorClass{Name: "bit_hamming_ops", AccessMethods: both, Type: "bit"},
		extdef.OperatorClass{Name: "bit_jaccard_ops", AccessMethods: hnswOnly, Type: "bit"},
	)
}

// accessMethods returns the two index access methods, both of which map to Dolt's native vector index.
func accessMethods() []extdef.AccessMethod {
	return []extdef.AccessMethod{
		{
			Name:    "hnsw",
			Handler: "hnswhandler",
			Params: []extdef.AccessMethodParam{
				{Name: "m", Min: 2, Max: 100, Default: 16},
				{Name: "ef_construction", Min: 4, Max: 1000, Default: 64},
			},
			CheckParams: func(params map[string]int64) error {
				if params["ef_construction"] < 2*params["m"] {
					return errors.New("ef_construction must be greater than or equal to 2 * m")
				}
				return nil
			},
		},
		{
			Name:    "ivfflat",
			Handler: "ivfflathandler",
			Params: []extdef.AccessMethodParam{
				{Name: "lists", Min: 1, Max: 32768, Default: 100},
			},
		},
	}
}

// aggregates returns the avg and sum aggregates for both dense types.
func aggregates() []extdef.Aggregate {
	aggs := make([]extdef.Aggregate, 0, 4)
	for _, typeName := range []string{"vector", "halfvec"} {
		aggs = append(aggs,
			extdef.Aggregate{
				Name:        "avg",
				Parameters:  params(typeName),
				Returns:     typeName,
				StateType:   "_float8",
				Transition:  typeName + "_accum",
				Final:       typeName + "_avg",
				Combine:     typeName + "_combine",
				InitCond:    "{0}",
				HasInitCond: true,
			},
			extdef.Aggregate{
				Name:       "sum",
				Parameters: params(typeName),
				Returns:    typeName,
				StateType:  typeName,
				Transition: typeName + "_add",
				Combine:    typeName + "_add",
			},
		)
	}
	return aggs
}
