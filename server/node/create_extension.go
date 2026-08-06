// Copyright 2025 Dolthub, Inc.
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

package node

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	coreextensions "github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/server/types"
)

// CreateExtension implements CREATE EXTENSION.
type CreateExtension struct {
	Name        string
	IfNotExists bool
	SchemaName  string
	Version     string
	Cascade     bool
}

var _ sql.ExecSourceRel = (*CreateExtension)(nil)
var _ vitess.Injectable = (*CreateExtension)(nil)

// NewCreateExtension returns a new *CreateExtension.
func NewCreateExtension(name string, ifNotExists bool, schemaName string, version string, cascade bool) *CreateExtension {
	return &CreateExtension{
		Name:        name,
		IfNotExists: ifNotExists,
		SchemaName:  schemaName,
		Version:     version,
		Cascade:     cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateExtension) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateExtension) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	if extCollection.HasLoadedExtension(ctx, id.NewExtension(c.Name)) {
		if c.IfNotExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`extension "%s" already exists`, c.Name)
	}
	// TODO: install the extensions named by Control.Requires, once an emulated extension declares any
	ext, err := extensions.Get(c.Name)
	if err != nil {
		return nil, err
	}

	schemaName, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	if err = (extensionObjects{ext: ext, schemaName: schemaName}).materialize(ctx, typColl); err != nil {
		return nil, err
	}
	err = extCollection.AddLoadedExtension(ctx, coreextensions.Extension{
		ExtName:     id.NewExtension(c.Name),
		Namespace:   id.NewNamespace(schemaName),
		Relocatable: ext.Control.Relocatable,
		Version:     ext.Control.DefaultVersion,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateExtension) String() string {
	return "CREATE EXTENSION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateExtension) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateExtension) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}

// extensionObjects materializes the objects that an extension declares.
type extensionObjects struct {
	ext        *extdef.Extension
	schemaName string
}

// materialize writes every object that the extension declares.
func (e extensionObjects) materialize(ctx *sql.Context, typColl *typecollection.TypeCollection) error {
	if len(e.ext.Routines) == 0 {
		return nil
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	for _, routine := range e.ext.Routines {
		returnType, err := e.typeID(ctx, typColl, routine.Returns)
		if err != nil {
			return err
		}
		paramTypes, err := e.parameterTypes(ctx, typColl, routine.Parameters)
		if err != nil {
			return err
		}
		allParams := make([]procedures.Parameter, len(routine.Parameters))
		for i, param := range routine.Parameters {
			allParams[i] = procedures.Parameter{Name: param.Name, Type: paramTypes[i]}
		}
		err = funcCollection.AddFunction(ctx, functions.Function{
			ID:                 id.NewFunction(e.schemaName, routine.Name, paramTypes...),
			ReturnType:         returnType,
			AllParams:          allParams,
			IsNonDeterministic: true,
			Strict:             routine.Strict,
			ExtensionName:      e.ext.Name,
			ExtensionSymbol:    routine.Symbol,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// typeID returns the type ID matching the given name.
func (e extensionObjects) typeID(ctx *sql.Context, typColl *typecollection.TypeCollection, name string) (id.Type, error) {
	_, typeID, err := typColl.ResolveName(ctx, doltdb.TableName{Name: name})
	if err != nil {
		return id.NullType, err
	}
	if !typeID.IsValid() {
		return id.NullType, types.ErrTypeDoesNotExist.New(name)
	}
	return id.Type(typeID), nil
}

// parameterTypes returns the type IDs of the given parameters.
func (e extensionObjects) parameterTypes(ctx *sql.Context, typColl *typecollection.TypeCollection, params []extdef.Parameter) ([]id.Type, error) {
	paramTypes := make([]id.Type, len(params))
	for i, param := range params {
		paramType, err := e.typeID(ctx, typColl, param.Type)
		if err != nil {
			return nil, err
		}
		paramTypes[i] = paramType
	}
	return paramTypes, nil
}
