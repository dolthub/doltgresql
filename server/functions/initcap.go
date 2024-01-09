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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// initcap represents the PostgreSQL function of the same name.
var initcap = Function{
	Name:      "initcap",
	Overloads: []interface{}{initcap_string},
}

// initcap_string is one of the overloads of initcap.
func initcap_string(text StringType) (StringType, error) {
	if text.IsNull {
		return StringType{IsNull: true}, nil
	}
	return StringType{Value: cases.Title(language.English).String(text.Value)}, nil
}
