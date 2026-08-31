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
	"bytes"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initFormatType registers the functions to the catalog.
func initFormatType() {
	framework.RegisterFunction(format_type)
}

// format_type represents the PostgreSQL system information function.
var format_type = framework.Function2{
	Name:       "format_type",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int32},
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		toid := id.Cache().ToOID(val1.(id.Id))
		if t, ok := types.OidToType[oid.Oid(toid)]; ok {
			if val2 == nil {
				return t.SQLStandardName(), nil
			} else {
				return t.SQLStandardNameWithTypmod(true, int(val2.(int32))), nil
			}
		}
		typ, err := getDoltgresTypeFromId(ctx, val1.(id.Id))
		if err != nil {
			if pgtypes.ErrTypeDoesNotExist.Is(err) {
				return "???", nil
			}
			return nil, err
		}
		return formatUserDefinedType(ctx, typ, val2)
	},
}

// formatUserDefinedType renders a catalog type using PostgreSQL's generic type-name and typmod rules.
func formatUserDefinedType(ctx *sql.Context, typ *pgtypes.DoltgresType, typmodValue any) (string, error) {
	isArray := typ.IsArrayType()
	if isArray {
		typ = typ.ArrayBaseType()
	}

	name, err := visibleTypeName(ctx, typ.ID)
	if err != nil {
		return "", err
	}
	if typmodValue != nil && typmodValue.(int32) >= 0 {
		typmod := typmodValue.(int32)
		if typ.ModOutFunc == 0 {
			name = fmt.Sprintf("%s(%d)", name, typmod)
		} else {
			modifier, err := typ.TypModOut(ctx, typmod)
			if err != nil {
				return "", err
			}
			name += modifier
		}
	}
	if isArray {
		name += "[]"
	}
	return name, nil
}

// visibleTypeName returns a quoted type name, schema-qualifying it when an unqualified reference would resolve elsewhere.
func visibleTypeName(ctx *sql.Context, typID id.Type) (string, error) {
	typCol, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return "", err
	}
	searchPath, err := core.SearchPath(ctx)
	if err != nil {
		return "", err
	}
	for _, schema := range searchPath {
		candidate, err := typCol.GetType(ctx, id.NewType(schema, typID.TypeName()))
		if err != nil {
			return "", err
		}
		if candidate != nil {
			if candidate.ID == typID {
				return quoteTypeIdentifier(typID.TypeName()), nil
			}
			break
		}
	}
	return quoteTypeIdentifier(typID.SchemaName()) + "." + quoteTypeIdentifier(typID.TypeName()), nil
}

// quoteTypeIdentifier quotes a type or schema name when PostgreSQL would not accept it as a bare identifier.
func quoteTypeIdentifier(name string) string {
	var buf bytes.Buffer
	lex.EncodeRestrictedSQLIdent(&buf, name, 0)
	return buf.String()
}
