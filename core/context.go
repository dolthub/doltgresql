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
	"maps"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/triggers"
	"github.com/dolthub/doltgresql/core/typecollection"
)

// contextValues contains a set of cached data passed alongside the context. This data is considered temporary
// and may be refreshed at any point, including during the middle of a query. Callers should not assume that
// data stored in contextValues is persisted, and other types of data should not be added to contextValues.
type contextValues struct {
	seqs  map[string]*sequences.Collection
	types map[string]*typecollection.TypeCollection
	funcs map[string]*functions.Collection
	procs map[string]*procedures.Collection
	trigs map[string]*triggers.Collection
	exts  map[string]*extensions.Collection
	casts map[string]*casts.Collection

	pgCatalogCache any
	runner         sql.StatementRunner

	// cache the dateOutputFormat, this is refreshed on SET
	dateOutputFormat string
}

// getContextValues accesses the contextValues in the given context. If the context does not have a contextValues, then
// it creates one and adds it to the context.
func getContextValues(ctx *sql.Context) (*contextValues, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
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

// ClearContextValues clears all context values. This is primarily for operations that are directly called from Dolt, as
// Dolt does not have the Doltgres concept of context values. Care must be taken to ensure that intermediate state
// written to the context values are not overwritten.
func ClearContextValues(ctx *sql.Context) {
	sess := dsess.DSessFromSess(ctx.Session)
	if sess.DoltgresSessObj != nil {
		// We want to persist the runner between clears since it's static
		var runner sql.StatementRunner
		if cv, ok := sess.DoltgresSessObj.(*contextValues); ok {
			runner = cv.runner
		}
		sess.DoltgresSessObj = &contextValues{
			runner: runner,
		}
	}
}

// GetRootFromContext returns the working session's root from the context, along with the session.
func GetRootFromContext(ctx *sql.Context) (*dsess.DoltSession, *RootValue, error) {
	return getRootFromContextForDatabase(ctx, "")
}

// getRootFromContextForDatabase returns the working session's root from the context for a specific database, along with the session.
func getRootFromContextForDatabase(ctx *sql.Context, database string) (*dsess.DoltSession, *RootValue, error) {
	session := dsess.DSessFromSess(ctx.Session)

	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	state, ok, err := session.LookupDbState(ctx, database)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, sql.ErrDatabaseNotFound.New(database)
	}
	// Some databases (e.g. Dolt's synthetic dolt_cluster system database) aren't backed by a Doltgres *RootValue
	// and never accumulate Doltgres-specific root object state (sequences, types, etc.), so there's nothing to
	// return here.
	root, ok := state.WorkingRoot().(*RootValue)
	if !ok {
		return session, nil, nil
	}
	return session, root, nil
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

// SetRunnerOnContext sets the given runner within the context values.
func SetRunnerOnContext(ctx *sql.Context, runner sql.StatementRunner) error {
	if runner == nil {
		return nil
	}
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	cv.runner = runner
	return nil
}

// GetRunnerFromContext returns the sql.StatementRunner from within the context.
func GetRunnerFromContext(ctx *sql.Context) (sql.StatementRunner, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	return cv.runner, nil
}

// GetDoltTableFromContext returns the Dolt table from the context. Returns nil if no table was found.
func GetDoltTableFromContext(ctx *sql.Context, tableName doltdb.TableName) (*doltdb.Table, error) {
	_, root, err := GetRootFromContext(ctx)
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
		searchPath, err = SearchPath(ctx)
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

// SearchPath returns the effective schema search path for the current session
func SearchPath(ctx *sql.Context) ([]string, error) {
	path, err := resolve.SearchPath(ctx)
	if err != nil {
		return nil, err
	}

	// pg_catalog is *always* implicitly part of the search path as the first element, unless it's specifically
	// included later. This allows users to override built-in names with user-defined names, but they have to
	// opt in to that behavior.
	hasPgCatalog := false
	for _, schema := range path {
		if schema == "pg_catalog" {
			hasPgCatalog = true
			break
		}
	}

	if !hasPgCatalog {
		path = append([]string{"pg_catalog"}, path...)
	}
	return path, nil
}

// GetExtensionsCollectionFromContext returns the extensions collection from the given context. Will always return a
// collection if no error is returned.
func GetExtensionsCollectionFromContext(ctx *sql.Context, database string) (*extensions.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.exts == nil {
		cv.exts = make(map[string]*extensions.Collection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.exts[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.exts[database], err = extensions.LoadExtensions(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.exts[database], nil
}

// GetFunctionsCollectionFromContext returns the functions collection from the given context. Will always return a
// collection if no error is returned.
func GetFunctionsCollectionFromContext(ctx *sql.Context, database string) (*functions.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.funcs == nil {
		cv.funcs = make(map[string]*functions.Collection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.funcs[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.funcs[database], err = functions.LoadFunctions(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.funcs[database], nil
}

// GetProceduresCollectionFromContext returns the procedures collection from the given context. Will always return a
// collection if no error is returned.
func GetProceduresCollectionFromContext(ctx *sql.Context, database string) (*procedures.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.procs == nil {
		cv.procs = make(map[string]*procedures.Collection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.procs[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.procs[database], err = procedures.LoadProcedures(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.procs[database], nil
}

// GetSequencesCollectionFromContext returns the given sequence collection from the context for the database
// named. If no database is provided, the context's current database is used.
// Will always return a collection if no error is returned.
func GetSequencesCollectionFromContext(ctx *sql.Context, database string) (*sequences.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.seqs == nil {
		cv.seqs = make(map[string]*sequences.Collection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.seqs[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.seqs[database], err = sequences.LoadSequences(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.seqs[database], nil
}

// GetTriggersCollectionFromContext returns the triggers collection from the given context. Will always return a
// collection if no error is returned.
func GetTriggersCollectionFromContext(ctx *sql.Context, database string) (*triggers.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.trigs == nil {
		cv.trigs = make(map[string]*triggers.Collection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.trigs[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.trigs[database], err = triggers.LoadTriggers(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.trigs[database], nil
}

// GetTypesCollectionFromContext returns the given type collection from the context.
// Will always return a collection if no error is returned.
func GetTypesCollectionFromContext(ctx *sql.Context, database string) (*typecollection.TypeCollection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.types == nil {
		cv.types = make(map[string]*typecollection.TypeCollection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.types[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.types[database], err = typecollection.LoadTypes(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.types[database], nil
}

// GetCastsCollectionFromContext returns the given casts collection from the context.
// Will always return a collection if no error is returned.
func GetCastsCollectionFromContext(ctx *sql.Context, database string) (*casts.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.casts == nil {
		cv.casts = make(map[string]*casts.Collection)
	}
	if len(database) == 0 {
		database = ctx.GetCurrentDatabase()
	}
	if cv.casts[database] == nil {
		_, root, err := getRootFromContextForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		cv.casts[database], err = casts.LoadCasts(ctx, root)
		if err != nil {
			return nil, err
		}
	}
	return cv.casts[database], nil
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

	// We need to update the root for all databases used by this context. This logic parallels what happens during
	// transaction commit in the dolt/sqle layer, where we check each branch state to see if it's dirty
	for _, db := range databasesInContext(ctx, cv) {
		err := updateSessionRootForDatabase(ctx, db, cv)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetDateStyleOutputFormat returns the cached DateOutputFormat
func GetDateStyleOutputFormat(ctx *sql.Context) (string, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return "", err
	}
	return cv.dateOutputFormat, nil
}

// SetDateStyleOutputFormat cached the provided dateOutputFormat
func SetDateStyleOutputFormat(ctx *sql.Context, dateOutputFormat string) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	cv.dateOutputFormat = dateOutputFormat
	return nil
}

// updateSessionRootForDatabase updates the root for all changes made to root object collections within the context
// values.
func updateSessionRootForDatabase(ctx *sql.Context, db string, cv *contextValues) error {
	session, root, err := getRootFromContextForDatabase(ctx, db)
	if err != nil {
		return err
	}

	newRoot := root
	if cv.seqs != nil && cv.seqs[db] != nil {
		retRoot, err := cv.seqs[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.seqs, db)
	}

	if cv.funcs != nil && cv.funcs[db] != nil && cv.funcs[db].DiffersFrom(ctx, root) {
		retRoot, err := cv.funcs[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.funcs, db)
	}

	if cv.procs != nil && cv.procs[db] != nil && cv.procs[db].DiffersFrom(ctx, root) {
		retRoot, err := cv.procs[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.procs, db)
	}

	if cv.trigs != nil && cv.trigs[db] != nil && cv.trigs[db].DiffersFrom(ctx, root) {
		retRoot, err := cv.trigs[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.trigs, db)
	}

	if cv.exts != nil && cv.exts[db] != nil && cv.exts[db].DiffersFrom(ctx, root) {
		retRoot, err := cv.exts[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.exts, db)
	}

	if cv.types != nil && cv.types[db] != nil {
		retRoot, err := cv.types[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.types, db)
	}

	if cv.casts != nil && cv.casts[db] != nil && cv.casts[db].DiffersFrom(ctx, root) {
		retRoot, err := cv.casts[db].UpdateRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		newRoot = retRoot.(*RootValue)
		delete(cv.casts, db)
	}

	// Setting the session working root doesn't do a check to see if anything actually changed or not before marking that
	// branch state dirty, and dolt only allows a single dirty working set per commit. So it's important here to only
	// update the session root if something actually changed for that db.
	if err, rootChanged := rootValueChanged(newRoot, root); rootChanged {
		if err = session.SetWorkingRoot(ctx, db, newRoot); err != nil {
			// TODO: We need a way to see if the session has a writeable working root
			// (new interface method on session probably), and avoid setting it if so
			if errors.Is(err, doltdb.ErrOperationNotSupportedInDetachedHead) {
				return nil
			}
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// rootValueChanged returns whether the new root value is different from the old one
func rootValueChanged(newRoot *RootValue, root *RootValue) (error, bool) {
	if newRoot == root {
		return nil, false
	}

	newHash, err := newRoot.HashOf()
	if err != nil {
		return err, false
	}

	oldHash, err := root.HashOf()
	if err != nil {
		return err, false
	}

	if newHash == oldHash {
		return nil, false
	}

	return nil, true
}

// databasesInContext returns all databases found within the context values.
func databasesInContext(ctx *sql.Context, cv *contextValues) []string {
	dbs := make(map[string]struct{})
	if cv.seqs != nil {
		for db := range cv.seqs {
			dbs[db] = struct{}{}
		}
	}
	currentDb := ctx.GetCurrentDatabase()
	if len(currentDb) > 0 {
		dbs[currentDb] = struct{}{}
	}

	return slices.Sorted(maps.Keys(dbs))
}

// clear removes the collection from the cache.
func (cv *contextValues) clear(objID objinterface.RootObjectID) {
	switch objID {
	case objinterface.RootObjectID_None:
		// Nothing to cache with this
	case objinterface.RootObjectID_Sequences:
		cv.seqs = nil
	case objinterface.RootObjectID_Types:
		cv.types = nil
	case objinterface.RootObjectID_Functions:
		cv.funcs = nil
	case objinterface.RootObjectID_Triggers:
		cv.trigs = nil
	case objinterface.RootObjectID_Extensions:
		// We don't cache these
	case objinterface.RootObjectID_Conflicts:
		// We don't cache these
	case objinterface.RootObjectID_Procedures:
		cv.procs = nil
	case objinterface.RootObjectID_Casts:
		cv.casts = nil
	default:
		panic("unhandled context clear object ID")
	}
}
