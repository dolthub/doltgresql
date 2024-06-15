// Copyright 2024 Dolthub, Inc.
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

package binary

// Init initializes all binary operators in this package.
func Init() {
	initBinaryBitAnd()
	initBinaryBitOr()
	initBinaryBitXor()
	initBinaryDivide()
	initBinaryEqual()
	initBinaryGreaterOrEqual()
	initBinaryGreaterThan()
	initBinaryLessOrEqual()
	initBinaryLessThan()
	initBinaryMinus()
	initBinaryMod()
	initBinaryMultiply()
	initBinaryNotEqual()
	initBinaryPlus()
	initBinaryShiftLeft()
	initBinaryShiftRight()
	initJSON()
}
