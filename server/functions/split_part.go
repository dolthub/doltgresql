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

package functions

import (
	"fmt"
	"strings"

	"github.com/dolthub/doltgresql/utils"
)

// split_part represents the PostgreSQL function of the same name.
var split_part = Function{
	Name:      "split_part",
	Overloads: []interface{}{split_part_string_string_int},
}

// split_part_string is one of the overloads of split_part.
func split_part_string_string_int(str StringType, delimiter StringType, n IntegerType) (StringType, error) {
	if str.IsNull || delimiter.IsNull || n.IsNull {
		return StringType{IsNull: true}, nil
	}
	if n.Value == 0 {
		return StringType{}, fmt.Errorf("field position must not be zero")
	}
	parts := strings.Split(str.Value, delimiter.Value)
	if int(utils.Abs(n.Value)) > len(parts) {
		return StringType{Value: ""}, nil
	}
	if n.Value > 0 {
		return StringType{Value: parts[n.Value-1]}, nil
	} else {
		return StringType{Value: parts[int64(len(parts))+n.Value]}, nil
	}
}
