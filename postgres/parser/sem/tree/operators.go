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

// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "github.com/cockroachdb/errors"

// This file implements the generation of unique names for every
// operator overload.
//
// The historical first purpose of generating these names is to be used
// as telemetry keys, for feature usage reporting.

// Detailed counter name generation follows.
//
// We pre-allocate the counter objects upfront here and later use
// Inc(), to avoid the hash map lookup in telemetry.Count upon type
// checking every scalar operator node.

// The logic that follows is also associated with a related feature in
// PostgreSQL, which may be implemented by CockroachDB in the future:
// exposing all the operators as unambiguous, non-overloaded built-in
// functions.  For example, in PostgreSQL, one can use `SELECT
// int8um(123)` to apply the int8-specific unary minus operator.
// This feature can be considered in the future for two reasons:
//
// 1. some pg applications may simply require the ability to use the
//    pg native operator built-ins. If/when this compatibility is
//    considered, care should be taken to tweak the string maps below
//    to ensure that the operator names generated here coincide with
//    those used in the postgres library.
//
// 2. since the operator built-in functions are non-overloaded, they
//    remove the requirement to disambiguate the type of operands
//    with the ::: (annotate_type) operator. This may be useful
//    to simplify/accelerate the serialization of scalar expressions
//    in distsql.
//

func init() {
	// Label the unary operators.
	for op := range UnaryOps {
		if int(op) >= len(unaryOpName) || unaryOpName[op] == "" {
			panic(errors.AssertionFailedf("missing name for operator %q", op.String()))
		}
	}

	// Label the comparison operators.
	for op := range CmpOps {
		if int(op) >= len(comparisonOpName) || comparisonOpName[op] == "" {
			panic(errors.AssertionFailedf("missing name for operator %q", op.String()))
		}
	}

	// Label the binary operators.
	for op := range BinOps {
		if int(op) >= len(binaryOpName) || binaryOpName[op] == "" {
			panic(errors.AssertionFailedf("missing name for operator %q", op.String()))
		}
	}
}
