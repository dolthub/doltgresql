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
	md5_package "crypto/md5"
	"fmt"
)

// md5 represents the PostgreSQL function of the same name.
var md5 = Function{
	Name:      "md5",
	Overloads: []interface{}{md5_string},
}

// md5_string is one of the overloads of md5.
func md5_string(text StringType) (StringType, error) {
	if text.IsNull {
		return StringType{IsNull: true}, nil
	}
	return StringType{Value: fmt.Sprintf("%x", md5_package.Sum([]byte(text.Value)))}, nil
}
