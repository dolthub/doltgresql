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

package pgcatalog

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgProcName is a constant to the pg_proc name.
const PgProcName = "pg_proc"

// InitPgProc handles registration of the pg_proc handler.
func InitPgProc() {
	tables.AddHandler(PgCatalogName, PgProcName, PgProcHandler{})
}

// PgProcHandler is the handler for the pg_proc table.
type PgProcHandler struct{}

// pgSequence represents a row in the pg_proc table
type pgProc struct {
	oid        id.Id  // oid
	name       string // proname
	schemaOid  id.Id  // pronamespace
	variadic   id.Id  // provariadic
	kind       string // prokind
	strict     bool   // proisstrict
	retSet     bool   // proretset
	volatile   string // provolatile
	nArgs      int16  // pronargs
	nArgDefs   int16  // pronargdefaults
	retTyp     id.Id  // prorettype
	argTypes   any    // proargtypes
	allArgTyps any    // proallargtypes
	argModes   any    // proargmodes
	argNames   any    // proargnames
	src        string // prosrc
	// TODO: Fill in the rest of the pg_proc columns
}

var _ tables.Handler = PgProcHandler{}

// Name implements the interface tables.Handler.
func (p PgProcHandler) Name() string {
	return PgProcName
}

// RowIter implements the interface tables.Handler.
func (p PgProcHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	cache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if cache.procs == nil {
		err = cachePgProcs(ctx, cache)
		if err != nil {
			return nil, err
		}
	}

	return &pgProcRowIter{
		procs: cache.procs,
		idx:   0,
	}, nil
}

// cachePgProcs caches the pg_proc data for the current database in the session.
func cachePgProcs(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var pprocs []*pgProc

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		// TODO: add built-in functions
		Function: func(ctx *sql.Context, schema functions.ItemSchema, f functions.ItemFunction) (cont bool, err error) {
			variadic := id.Null
			if f.Item.Variadic {
				// TODO not implemented yet
				variadic = id.Null
			}

			nArgs := int16(len(f.Item.ParameterTypes))
			nArgDefs := int16(0)
			retSet := false
			if f.Item.ReturnType.IsValid() && f.Item.ReturnType.TypeName() == "record" {
				retSet = true
			}

			var (
				kind               = "f" // a for an aggregate function, or w for a window function
				argTypes, argNames any
				types, names       []any
			)

			hasNonEmtpyArgName := false
			for i, typ := range f.Item.ParameterTypes {
				if f.Item.ParameterDefaults[i] != "" {
					nArgDefs += 1
				}
				if f.Item.ParameterNames[i] != "" {
					names = append(names, f.Item.ParameterNames[i])
				}
				types = append(types, typ.AsId())
				if f.Item.ParameterNames[i] != "" {
					hasNonEmtpyArgName = true
				}
				names = append(names, f.Item.ParameterNames[i])
			}

			if len(types) > 0 {
				argTypes = types
			}
			if hasNonEmtpyArgName && len(names) > 0 {
				argNames = names
			}

			var volatile = "i" // immutable
			if f.Item.IsNonDeterministic {
				volatile = "v" // volatile
			}

			pprocs = append(pprocs, &pgProc{
				oid:        f.OID.AsId(),
				name:       f.Item.ID.FunctionName(),
				schemaOid:  schema.OID.AsId(),
				variadic:   variadic,
				kind:       kind,
				strict:     f.Item.Strict,
				retSet:     retSet,
				volatile:   volatile,
				nArgs:      nArgs,
				nArgDefs:   nArgDefs,
				retTyp:     f.Item.ReturnType.AsId(),
				argTypes:   argTypes,
				allArgTyps: nil,
				argModes:   nil,
				argNames:   argNames,
				src:        f.Item.SQLDefinition,
			})
			return true, nil
		},
		Procedure: func(ctx *sql.Context, schema functions.ItemSchema, p functions.ItemProcedure) (cont bool, err error) {
			nArgs := int16(len(p.Item.ParameterTypes))
			nArgDefs := int16(0)

			var (
				// argTypes includes only input arguments (including INOUT and VARIADIC arguments)
				argTypes any
				// argAllTypes includes all arguments (including OUT and INOUT arguments);
				// however, if all the arguments are IN arguments, this field will be null.
				argAllTypes any
				argNames    any
			)

			var types, allTypes, names []any
			hasNonINArg := false
			hasNonEmtpyArgName := false
			for i, typ := range p.Item.ParameterTypes {
				switch p.Item.ParameterModes[i] {
				case procedures.ParameterMode_IN:
					types = append(types, typ.AsId())
				case procedures.ParameterMode_VARIADIC, procedures.ParameterMode_INOUT:
					types = append(types, typ.AsId())
					hasNonINArg = true
				case procedures.ParameterMode_OUT:
					hasNonINArg = true
				}
				if p.Item.ParameterDefaults[i] != "" {
					nArgDefs += 1
				}
				if p.Item.ParameterNames[i] != "" {
					hasNonEmtpyArgName = true
				}
				allTypes = append(allTypes, typ.AsId())
				names = append(names, p.Item.ParameterNames[i])
			}

			if len(types) > 0 {
				argTypes = types
			}
			if hasNonINArg && len(allTypes) > 0 {
				argAllTypes = allTypes
			}
			if hasNonEmtpyArgName && len(names) > 0 {
				argNames = names
			}

			pprocs = append(pprocs, &pgProc{
				oid:        p.OID.AsId(),
				name:       p.Item.ID.ProcedureName(),
				schemaOid:  schema.OID.AsId(),
				variadic:   id.Null,
				kind:       "p",
				strict:     false,
				retSet:     false,
				volatile:   "v", // volatile
				nArgs:      nArgs,
				nArgDefs:   nArgDefs,
				retTyp:     pgtypes.Void.ID.AsId(),
				argTypes:   argTypes,
				allArgTyps: argAllTypes,
				argModes:   nil,
				argNames:   argNames,
				src:        p.Item.SQLDefinition,
			})
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	pgCatalogCache.procs = pprocs
	return nil
}

// PkSchema implements the interface tables.Handler.
func (p PgProcHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgProcSchema,
		PkOrdinals: nil,
	}
}

// pgProcSchema is the schema for pg_proc.
var pgProcSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "pronamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prolang", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "procost", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prorows", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "provariadic", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prosupport", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgProcName}, // TODO: type regproc
	{Name: "prokind", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prosecdef", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proleakproof", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proisstrict", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proretset", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "provolatile", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proparallel", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "pronargs", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "pronargdefaults", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prorettype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proargtypes", Type: pgtypes.Oidvector, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proallargtypes", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgProcName},
	{Name: "proargmodes", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName}, // TODO: type char[]
	{Name: "proargnames", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName}, // TODO: collation C
	{Name: "proargdefaults", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgProcName},   // TODO: type pg_node_tree, collation C
	{Name: "protrftypes", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgProcName},
	{Name: "prosrc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgProcName}, // TODO: collation C
	{Name: "probin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgProcName},
	{Name: "prosqlbody", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgProcName},     // TODO: type pg_node_tree, collation C
	{Name: "proconfig", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName}, // TODO: collation C
	{Name: "proacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName},    // TODO: type aclitem[]
}

// pgProcRowIter is the sql.RowIter for the pg_proc table.
type pgProcRowIter struct {
	procs []*pgProc
	idx   int
}

var _ sql.RowIter = (*pgProcRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgProcRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.procs) {
		return nil, io.EOF
	}
	p := iter.procs[iter.idx]
	iter.idx++

	return sql.Row{
		p.oid,        // oid
		p.name,       // proname
		p.schemaOid,  // pronamespace
		id.Null,      // proowner
		id.Null,      // prolang
		float32(1),   // procost
		float32(0),   // prorows
		p.variadic,   // provariadic
		nil,          // prosupport
		p.kind,       // prokind
		false,        // prosecdef
		false,        // proleakproof
		p.strict,     // proisstrict
		p.retSet,     // proretset
		p.volatile,   // provolatile
		"u",          // proparallel // TODO: default to 'unsafe' for now
		p.nArgs,      // pronargs
		p.nArgDefs,   // pronargdefaults
		p.retTyp,     // prorettype
		p.argTypes,   // proargtypes
		p.allArgTyps, // proallargtypes
		p.argModes,   // proargmodes
		p.argNames,   // proargnames
		nil,          // proargdefaults
		nil,          // protrftypes
		p.src,        // prosrc
		nil,          // probin
		nil,          // prosqlbody
		nil,          // proconfig
		nil,          // proacl
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgProcRowIter) Close(ctx *sql.Context) error {
	return nil
}
