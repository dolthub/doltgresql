// Copyright 2023 Dolthub, Inc.
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

package main

import (
	"crypto/rand"
	"math"
	"math/big"
	"sort"
)

var (
	bigIntZero      = big.NewInt(0)
	bigIntOne       = big.NewInt(1)
	bigIntTwo       = big.NewInt(2)
	bigIntMaxInt64  = big.NewInt(math.MaxInt64)
	bigIntMaxUint64 = new(big.Int).Add(new(big.Int).Mul(bigIntMaxInt64, bigIntTwo), bigIntOne)
)

// GenerateRandomInts generates a slice of random integers, with each integer ranging from [0, max). The returned slice
// will be sorted from smallest to largest. If count <= 0 or max <= 0, then they will be set to 1.
func GenerateRandomInts(count int64, max *big.Int) (randInts []*big.Int, err error) {
	if count <= 0 {
		count = 1
	}
	if max.Cmp(bigIntZero) == -1 {
		max = bigIntOne
	}
	randInts = make([]*big.Int, count)
	for i := range randInts {
		randInts[i], err = rand.Int(rand.Reader, max)
		if err != nil {
			return nil, err
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
