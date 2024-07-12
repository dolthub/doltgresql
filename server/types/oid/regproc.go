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

package oid

import (
	"fmt"
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"
)

// regproc_IoInput is the implementation for IoInput that avoids circular dependencies by being declared in a separate
// package.
func regproc_IoInput(ctx *sql.Context, input string) (uint32, error) {
	// If the string just represents a number, then we return it.
	if parsedOid, err := strconv.ParseUint(input, 10, 32); err == nil {
		return uint32(parsedOid), nil
	}
	sections, err := ioInputSections(input)
	if err != nil {
		return 0, err
	}
	if err = regproc_IoInputValidation(ctx, input, sections); err != nil {
		return 0, err
	}
	switch len(sections) {
	case 1:
		// TODO: handle procedures, aggregate functions, and window functions
		name := sections[0]
		oid := uint32(0)
		err = IterateCurrentDatabase(ctx, Callbacks{
			Function: func(ctx *sql.Context, function ItemFunction) (cont bool, err error) {
				if function.Item.FunctionName() == name {
					// TODO: this should error for overloaded functions
					oid = function.OID
					return false, nil
				}
				return true, nil
			},
		})
		if err != nil || oid != 0 {
			return oid, err
		}
		return 0, fmt.Errorf(`"function "%s" does not exist"`, input)
	default:
		return 0, fmt.Errorf("regproc failed validation")
	}
}

// regproc_IoOutput is the implementation for IoOutput that avoids circular dependencies by being declared in a separate
// package.
func regproc_IoOutput(ctx *sql.Context, oid uint32) (string, error) {
	output := strconv.FormatUint(uint64(oid), 10)
	err := RunCallback(ctx, oid, Callbacks{
		Function: func(ctx *sql.Context, function ItemFunction) (cont bool, err error) {
			output = function.Item.FunctionName()
			return false, nil
		},
	})
	return output, err
}

// regproc_IoInputValidation handles the validation for the parsed sections in regproc_IoInput.
func regproc_IoInputValidation(ctx *sql.Context, input string, sections []string) error {
	switch len(sections) {
	case 1:
		return nil
	case 3:
		if sections[1] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return fmt.Errorf("functions are not yet implemented in terms of the schema")
	case 5:
		if sections[1] != "." || sections[3] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return fmt.Errorf("cross-database references are not implemented: %s", input)
	case 7:
		if sections[1] != "." || sections[3] != "." || sections[5] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return fmt.Errorf("improper qualified name (too many dotted names): %s", input)
	default:
		return fmt.Errorf("invalid name syntax")
	}
}
