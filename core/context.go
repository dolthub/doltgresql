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

package core

import (
	"errors"
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/types"
)

// contextValues contains a set of objects that will be passed alongside the context.
type contextValues struct {
	collection *sequences.Collection
	types      *types.TypeCollection
}

// getContextValues accesses the contextValues in the given context. If the context does not have a contextValues, then
// it creates one and adds it to the context.
func getContextValues(ctx *sql.Context) (*contextValues, error) {
	sess := dsess.DSessFromSess(ctx.Session)
	if sess.DoltgresSessObj == nil {
		cv := &contextValues{}
		sess.DoltgresSessObj = cv
		return cv, nil
	}
	cv, ok := sess.DoltgresSessObj.(*contextValues)
	if !ok {
		return nil, fmt.Errorf("context contains an unknown values struct of type: %T", sess.DoltgresSessObj)
	}
	return cv, nil
}

// getRootFromContext returns the working session's root from the context, along with the session.
func getRootFromContext(ctx *sql.Context) (*dsess.DoltSession, *RootValue, error) {
	session := dsess.DSessFromSess(ctx.Session)
	// Does this handle the current schema as well?
	state, ok, err := session.LookupDbState(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, fmt.Errorf("cannot find the database while fetching root from context")
	}
	return session, state.WorkingRoot().(*RootValue), nil
}

// GetDoltTableFromContext returns the Dolt table from the context. Returns nil if no table was found.
func GetDoltTableFromContext(ctx *sql.Context, tableName doltdb.TableName) (*doltdb.Table, error) {
	_, root, err := getRootFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var table *doltdb.Table
	if tableName.Schema == "" {
		_, table, _, err = resolve.Table(ctx, root, tableName.Name)
		if err != nil {
			return nil, err
		}
	} else {
		table, _, err = root.GetTable(ctx, tableName)
	}

	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetSqlDatabaseFromContext returns the database from the context. Uses the context's current database if an empty
// string is provided. Returns nil if the database was not found.
func GetSqlDatabaseFromContext(ctx *sql.Context, database string) (sql.Database, error) {
	session := dsess.DSessFromSess(ctx.Session)
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	db, err := session.Provider().Database(ctx, database)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, nil
		}
		return nil, err
	}
	return db, nil
}

// GetSqlTableFromContext returns the table from the context. Uses the context's current database if an empty database
// name is provided. Returns nil if no table was found.
func GetSqlTableFromContext(ctx *sql.Context, databaseName string, tableName doltdb.TableName) (sql.Table, error) {
	db, err := GetSqlDatabaseFromContext(ctx, databaseName)
	if err != nil || db == nil {
		return nil, err
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		// Fairly confident that Dolt only has database implementations that inherit sql.SchemaDatabase, so only GMS
		// databases may fail here (like information_schema). In this scenario, we expect that no schema will be passed,
		// so we'll special-case it here.
		if len(tableName.Schema) == 0 {
			tbl, ok, err := db.GetTableInsensitive(ctx, tableName.Name)
			if err != nil || !ok {
				return nil, err
			}
			return tbl, nil
		}
		return nil, nil
	}

	var searchPath []string
	if len(tableName.Schema) == 0 {
		// If a schema was not provided, then we'll use the search path
		searchPath, err = resolve.SearchPath(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		// A specific schema is given, so we'll only use that one for the search path
		searchPath = []string{tableName.Schema}
	}

	for _, schema := range searchPath {
		db, ok, err = schemaDb.GetSchema(ctx, schema)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		tbl, ok, err := db.GetTableInsensitive(ctx, tableName.Name)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		return tbl, nil
	}
	return nil, nil
}

// GetSequencesCollectionFromContext returns the given sequence collection from the context. Will always return a collection if
// no error is returned.
func GetSequencesCollectionFromContext(ctx *sql.Context) (*sequences.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.collection == nil {
		_, root, err := getRootFromContext(ctx)
		if err != nil {
			return nil, err
		}
		cv.collection, err = root.GetSequences(ctx)
		if err != nil {
			return nil, err
		}
	}
	return cv.collection, nil
}

// GetTypesCollectionFromContext returns the given domain collection from the context.
// Will always return a collection if no error is returned.
func GetTypesCollectionFromContext(ctx *sql.Context) (*types.TypeCollection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.types == nil {
		_, root, err := getRootFromContext(ctx)
		if err != nil {
			return nil, err
		}
		cv.types, err = root.GetTypes(ctx)
		if err != nil {
			return nil, err
		}
	}
	return cv.types, nil
}

// CloseContextRootFinalizer finalizes any changes persisted within the context by writing them to the working root.
// This should ONLY be called by the ContextRootFinalizer node.
func CloseContextRootFinalizer(ctx *sql.Context) error {
	sess := dsess.DSessFromSess(ctx.Session)
	if sess.DoltgresSessObj == nil {
		return nil
	}
	cv, ok := sess.DoltgresSessObj.(*contextValues)
	if !ok {
		return nil
	}
	if cv.collection == nil {
		return nil
	}
	session, root, err := getRootFromContext(ctx)
	if err != nil {
		return err
	}
	newRoot, err := root.PutSequences(ctx, cv.collection)
	if err != nil {
		return err
	}
	if newRoot != nil {
		if err = session.SetWorkingRoot(ctx, ctx.GetCurrentDatabase(), newRoot); err != nil {
			// TODO: We need a way to see if the session has a writeable working root
			// (new interface method on session probably), and avoid setting it if so
			if errors.Is(err, doltdb.ErrOperationNotSupportedInDetachedHead) {
				return nil
			}
			return err
		}
	}
	return nil
}
