// Copyright 2023-2024 Dolthub, Inc.
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

package utils

import (
	"crypto/rand"
	"math"
	"math/big"
	"sort"
)

var (
	BigIntZero      = big.NewInt(0)
	BigIntOne       = big.NewInt(1)
	BigIntTwo       = big.NewInt(2)
	BigIntMaxInt64  = big.NewInt(math.MaxInt64)
	BigIntMaxUint64 = new(big.Int).Add(new(big.Int).Mul(BigIntMaxInt64, BigIntTwo), BigIntOne)
)

// GenerateRandomInts generates a slice of random integers, with each integer ranging from [0, max). The returned slice
// will be sorted from smallest to largest. If count <= 0 or max <= 0, then they will be set to 1. If count >= max, then
// the returned slice will contain all incrementing integers [0, max).
func GenerateRandomInts(count int64, max *big.Int) (randInts []*big.Int, err error) {
	if count <= 0 {
		count = 1
	}
	if max.Cmp(BigIntZero) == -1 {
		max = BigIntOne
	}
	// If count >= max, then we'll just shortcut and add incrementing integers up to the max (not including the max)
	if big.NewInt(count).Cmp(max) >= 0 {
		max64 := max.Int64()
		randInts = make([]*big.Int, max64)
		for i := int64(0); i < max64; i++ {
			randInts[i] = big.NewInt(i)
		}
		return randInts, nil
	}

	randInts = make([]*big.Int, count)
	randIntSet := make(map[string]struct{}, count*2)
	for i := range randInts {
		for {
			randInts[i], err = rand.Int(rand.Reader, max)
			if err != nil {
				return nil, err
			}
			if _, ok := randIntSet[randInts[i].String()]; !ok {
				randIntSet[randInts[i].String()] = struct{}{}
				break
			}
		}
	}
	sort.Slice(randInts, func(i, j int) bool {
		return randInts[i].Cmp(randInts[j]) == -1
	})
	return randInts, nil
}

// GetPercentages converts the slice of numbers to percentages. The max defines the number that would equal 100%. All
// floats will be between [0.0, 100.0], unless the number is not between [0, max].
func GetPercentages(numbers []*big.Int, max *big.Int) []float64 {
	maxAsFloat, _ := max.Float64()
	percentages := make([]float64, len(numbers))
	for i, number := range numbers {
		numberAsFloat, _ := number.Float64()
		percentages[i] = (numberAsFloat / maxAsFloat) * 100.0
	}
	return percentages
}
