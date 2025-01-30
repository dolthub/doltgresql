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
	"regexp"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initSubstr registers the functions to the catalog.
func initSubstr() {
	framework.RegisterFunction(substr_text_int32)
	framework.RegisterFunction(substring_text_int32)
	framework.RegisterFunction(substr_text_int32_int32)
	framework.RegisterFunction(substring_text_int32_int32)
	framework.RegisterFunction(substring_text_text)
}

// substr_text_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var substr_text_int32 = framework.Function2{
	Name:       "substr",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32},
	Strict:     true,
	Callable:   substring_text_int32_fn,
}

// substring_text_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var substring_text_int32 = framework.Function2{
	Name:       "substring",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32},
	Strict:     true,
	Callable:   substring_text_int32_fn,
}

// substring_text_int32_fn is a helper function for substr_text_int32 and substring_text_int32.
func substring_text_int32_fn(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, str any, start any) (any, error) {
	runes := []rune(str.(string))
	if start.(int32) < 1 {
		start = int32(1)
	}
	// start is 1-indexed
	start = start.(int32) - int32(1)
	if int(start.(int32)) >= len(runes) {
		return "", nil
	}
	return string(runes[start.(int32):]), nil
}

// substr_text_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var substr_text_int32_int32 = framework.Function3{
	Name:       "substr",
	Return:     pgtypes.Text,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable:   substring_text_int32_int32_fn,
}

// substr_text_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var substring_text_int32_int32 = framework.Function3{
	Name:       "substring",
	Return:     pgtypes.Text,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable:   substring_text_int32_int32_fn,
}

// substring_text_int32_int32_fn is a helper function for substr_text_int32_int32.
func substring_text_int32_int32_fn(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, str any, startInt any, countInt any) (any, error) {
	start := startInt.(int32)
	count := countInt.(int32)
	runes := []rune(str.(string))
	if count < 0 {
		return nil, errors.Errorf("negative substring length not allowed")
	}
	// start is 1-indexed
	start--
	if start < 0 {
		count += start
		start = 0
	}
	if count <= 0 {
		return "", nil
	}
	if int(start) >= len(runes) {
		return "", nil
	} else if int64(start)+int64(count) > int64(len(runes)) {
		return string(runes[start:]), nil
	}
	return string(runes[start : start+count]), nil
}

// substr_text_text represents the PostgreSQL function of the same name, taking the same parameters.
// This is a form of the syntax `substring(string from pattern)`.
var substring_text_text = framework.Function2{
	Name:       "substring",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, str any, pattern any) (any, error) {
		re, err := regexp.Compile(pattern.(string))
		if err != nil {
			return nil, err
		}

		match := re.Find([]byte(str.(string)))

		if match == nil {
			return nil, nil
		}

		return string(match), nil
	},
}
