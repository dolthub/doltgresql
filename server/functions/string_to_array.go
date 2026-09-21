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

package functions

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initStringToArray registers the functions to the catalog.
func initStringToArray() {
	framework.RegisterFunction(string_to_array_text_text)
	framework.RegisterFunction(string_to_array_text_text_text)
}

// string_to_array_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var string_to_array_text_text = framework.Function2{
	Name:       "string_to_array",
	Return:     pgtypes.TextArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     false,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, str any, delimiter any) (any, error) {
		return stringToArray(ctx, str, delimiter, nil, false)
	},
}

// string_to_array_text_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var string_to_array_text_text_text = framework.Function3{
	Name:       "string_to_array",
	Return:     pgtypes.TextArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     false,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, str any, delimiter any, nullStr any) (any, error) {
		return stringToArray(ctx, str, delimiter, nullStr, nullStr != nil)
	},
}

// stringToArray implements the semantics of the PostgreSQL string_to_array function. If the input string is NULL, the
// result is NULL. If the delimiter is NULL, each character of the input string becomes a separate element. If the
// delimiter is an empty string, the entire input string is returned as a single-element array. An empty input string
// always produces an empty array. When hasNullStr is true, any field equal to nullStr is replaced with NULL.
func stringToArray(ctx *sql.Context, str any, delimiter any, nullStr any, hasNullStr bool) (any, error) {
	if str == nil {
		return nil, nil
	}
	inputStr, err := framework.UnwrapString(ctx, str)
	if err != nil {
		return nil, err
	}
	var nullStrStr string
	if hasNullStr {
		nullStrStr, err = framework.UnwrapString(ctx, nullStr)
		if err != nil {
			return nil, err
		}
	}

	var fields []string
	if delimiter == nil {
		// A NULL delimiter splits the string into individual characters
		for _, r := range inputStr {
			fields = append(fields, string(r))
		}
	} else {
		delimiterStr, dErr := framework.UnwrapString(ctx, delimiter)
		if dErr != nil {
			return nil, dErr
		}
		if len(inputStr) == 0 {
			// An empty input string always produces an empty array
			return []any{}, nil
		}
		if len(delimiterStr) == 0 {
			// An empty delimiter returns the entire input string as a single field
			fields = []string{inputStr}
		} else {
			fields = strings.Split(inputStr, delimiterStr)
		}
	}

	result := make([]any, len(fields))
	for i, field := range fields {
		if hasNullStr && field == nullStrStr {
			result[i] = nil
		} else {
			result[i] = field
		}
	}
	return result, nil
}
