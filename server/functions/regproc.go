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
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegproc registers the functions to the catalog.
func initRegproc() {
	framework.RegisterFunction(regprocin)
	framework.RegisterFunction(regprocout)
	framework.RegisterFunction(regprocrecv)
	framework.RegisterFunction(regprocsend)
}

// regprocin represents the PostgreSQL function of regproc type IO input.
var regprocin = framework.Function1{
	Name:       "regprocin",
	Return:     pgtypes.Regproc,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// If the string just represents a number, then we return it.
		input := val.(string)
		if parsedOid, err := strconv.ParseUint(input, 10, 32); err == nil {
			if internalID := id.Cache().ToInternal(uint32(parsedOid)); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewInternalOID(uint32(parsedOid)).Internal(), nil
		}
		sections, err := ioInputSections(input)
		if err != nil {
			return id.Null, err
		}
		if err = regproc_IoInputValidation(ctx, input, sections); err != nil {
			return id.Null, err
		}
		switch len(sections) {
		case 1:
			// TODO: handle procedures, aggregate functions, and window functions
			// TODO: this only handles built-in functions
			funcInterfaces := framework.Catalog[sections[0]]
			if len(funcInterfaces) == 1 {
				return funcInterfaces[0].InternalID(), nil
			}
			return id.Null, fmt.Errorf(`"function "%s" does not exist"`, input)
		default:
			return id.Null, fmt.Errorf("regproc failed validation")
		}
	},
}

// regprocout represents the PostgreSQL function of regproc type IO output.
var regprocout = framework.Function1{
	Name:       "regprocout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regproc},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(id.Internal)
		if input.Section() == id.Section_OID {
			return input.Segment(0), nil
		}
		return val.(id.Internal).Segment(1), nil
	},
}

// regprocrecv represents the PostgreSQL function of regproc type IO receive.
var regprocrecv = framework.Function1{
	Name:       "regprocrecv",
	Return:     pgtypes.Regproc,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return id.Internal(data), nil
	},
}

// regprocsend represents the PostgreSQL function of regproc type IO send.
var regprocsend = framework.Function1{
	Name:       "regprocsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regproc},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(id.Internal)), nil
	},
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
