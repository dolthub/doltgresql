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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/typecollection"
)

// contextValues contains a set of objects that will be passed alongside the context.
type contextValues struct {
	collection     *sequences.Collection
	types          *typecollection.TypeCollection
	funcs          *functions.Collection
	pgCatalogCache any

	// backend is the connection to the client, used to send postgres protocol messages back to the client
	// (e.g. when a RAISE statement sends a NOTICE to the client)
	backend *pgproto3.Backend
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
		return nil, errors.Errorf("context contains an unknown values struct of type: %T", sess.DoltgresSessObj)
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
		return nil, nil, errors.Errorf("cannot find the database while fetching root from context")
	}
	return session, state.WorkingRoot().(*RootValue), nil
}

// IsContextValid returns whether the context is valid for use with any of the functions in the package. If this is not
// false, then there's a high likelihood that the context is being used in a temporary scenario (such as setting up the
// database, etc.).
func IsContextValid(ctx *sql.Context) bool {
	if ctx == nil {
		return false
	}
	_, ok := ctx.Session.(*dsess.DoltSession)
	return ok
}

// GetPgCatalogCache returns a cache of data for pg_catalog tables. This function should only be used by
// pg_catalog table handlers. The catalog cache instance stores generated pg_catalog table data so that
// it only has to generated table data once per query.
//
// TODO: The return type here is currently any, to avoid a package import cycle. This could be cleaned up by
// introducing a new interface type, in a package that does not depend on core or pgcatalog packages.
func GetPgCatalogCache(ctx *sql.Context) (any, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	return cv.pgCatalogCache, nil
}

// SetPgCatalogCache sets |pgCatalogCache| as the catalog cache instance for this session.
//
// TODO: The input type here is currently any, to avoid a package import cycle. This could be cleaned up by
// introducing a new interface type, in a package that does not depend on core or pgcatalog packages.
func SetPgCatalogCache(ctx *sql.Context, pgCatalogCache any) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	cv.pgCatalogCache = pgCatalogCache
	return nil
}

// GetBackend returns the pgproto3 Backend instance stored in the specified |ctx|.
func GetBackend(ctx *sql.Context) (*pgproto3.Backend, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	return cv.backend, nil
}

// SetBackend sets the pgproto3 |backend| in the specified |ctx|, so that other parts of the engine can access it
// to send messages, such as Notices.
func SetBackend(ctx *sql.Context, backend *pgproto3.Backend) error {
	// TODO: Instead of exposing the Backend instance directly, if only support for sending Notices is needed,
	//       we could pass in a type that is specific for sending Notices.
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	cv.backend = backend
	return nil
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

// GetFunctionsCollectionFromContext returns the functions collection from the given context. Will always return a
// collection if no error is returned.
func GetFunctionsCollectionFromContext(ctx *sql.Context) (*functions.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.funcs == nil {
		_, root, err := getRootFromContext(ctx)
		if err != nil {
			return nil, err
		}
		cv.funcs, err = root.GetFunctions(ctx)
		if err != nil {
			return nil, err
		}
	}
	return cv.funcs, nil
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

// GetTypesCollectionFromContext returns the given type collection from the context.
// Will always return a collection if no error is returned.
func GetTypesCollectionFromContext(ctx *sql.Context) (*typecollection.TypeCollection, error) {
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
	newRoot, err = newRoot.PutFunctions(ctx, cv.funcs)
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
