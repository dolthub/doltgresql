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

package expression

// NewSomeExpr creates a new AnyExpr expression for SOME. SOME is synonymous with ANY.
func NewSomeExpr(subOperator string) *AnyExpr {
	return &AnyExpr{
		leftExpr:    nil,
		rightExpr:   nil,
		subOperator: subOperator,
		name:        "SOME",
	}
}
