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

package _go

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

func TestExtensionEmulation(t *testing.T) {
	registerTestExtension()
	RunScripts(t, []ScriptTest{
		{
			Name: "Declared types are created with their array type",
			SetUpScript: []string{
				"CREATE EXTENSION doltgres_test;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT typname, typtype, typcategory, typlen, typinput::text, typoutput::text FROM pg_type WHERE typname = 'dgtest_upper';`,
					Expected: []sql.Row{
						{"dgtest_upper", "b", "U", -1, "dgtest_upper_in", "dgtest_upper_out"},
					},
				},
				{
					Query:    `SELECT typname, typtype, typcategory FROM pg_type WHERE typname = '_dgtest_upper';`,
					Expected: []sql.Row{{"_dgtest_upper", "b", "A"}},
				},
				{
					Query:    "SELECT 'abc'::dgtest_upper;",
					Expected: []sql.Row{{"ABC"}},
				},
				{
					Query:    "SELECT 'MiXeD'::dgtest_upper;",
					Expected: []sql.Row{{"MIXED"}},
				},
			},
		},
		{
			Name: "Declared operators resolve for their operand types",
			SetUpScript: []string{
				"CREATE EXTENSION doltgres_test;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT 'abc'::dgtest_upper = 'ABC'::dgtest_upper;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT 'abc'::dgtest_upper = 'xyz'::dgtest_upper;",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT oprname, oprkind, oprcanhash, oprcanmerge FROM pg_operator WHERE oprcode::text = 'dgtest_upper_eq';`,
					Expected: []sql.Row{{"=", "b", "t", "t"}},
				},
				{ // A symbol that no built-in operator resolves through the same path
					Query:    "SELECT 'abc'::dgtest_upper <-> 'wxyz'::dgtest_upper;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT 'abc'::dgtest_upper <-> 'xyz'::dgtest_upper;",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "Declared casts convert to their target type",
			SetUpScript: []string{
				"CREATE EXTENSION doltgres_test;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ('abc'::dgtest_upper)::text || 'def';",
					Expected: []sql.Row{{"ABCdef"}},
				},
				{
					Query:    "SELECT length(('abcd'::dgtest_upper)::text);",
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name: "Declared aggregates accumulate through their transition and final functions",
			SetUpScript: []string{
				"CREATE EXTENSION doltgres_test;",
				"CREATE TABLE t1 (pk INTEGER PRIMARY KEY, v1 TEXT);",
				"INSERT INTO t1 VALUES (1, 'ab'), (2, 'cde'), (3, 'f');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT dgtest_charcount(v1) FROM t1;",
					Expected: []sql.Row{{"6"}},
				},
				{
					Query:    "SELECT dgtest_charcount(v1) FROM t1 WHERE pk = 2;",
					Expected: []sql.Row{{"3"}},
				},
				{ // The initial condition is the state when the aggregate sees no rows at all
					Query:    "SELECT dgtest_charcount(v1) FROM t1 WHERE pk = 0;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    `SELECT aggkind, agginitval, aggcombinefn::text, aggfinalfn::text FROM pg_aggregate WHERE aggtransfn::text = 'dgtest_charcount_transition';`,
					Expected: []sql.Row{{"n", "0", "dgtest_charcount_combine", "dgtest_charcount_final"}},
				},
			},
		},
		{
			Name: "Objects are created in the extension's target schema",
			SetUpScript: []string{
				"CREATE EXTENSION doltgres_test;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT n.nspname FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE t.typname = 'dgtest_upper';`,
					Expected: []sql.Row{{"public"}},
				},
				{
					Query:    "SELECT public.dgtest_upper_text('abc'::public.dgtest_upper);",
					Expected: []sql.Row{{"ABC"}},
				},
			},
		},
		{
			Name: "Extension objects are not available before the extension is created",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT 'abc'::dgtest_upper;",
					ExpectedErr: "unable to resolve type",
				},
				{
					Query:       "SELECT dgtest_charcount('abc');",
					ExpectedErr: "not found",
				},
				{
					Query:    "CREATE EXTENSION doltgres_test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT dgtest_charcount('abc');",
					Expected: []sql.Row{{"3"}},
				},
			},
		},
	})
}

// registerExtensionOnce ensures that registerTestExtension only runs once.
var registerExtensionOnce sync.Once

// registerTestExtension registers a dummy extension for testing, which declares one of every kind of object. Every test
// whose expected output includes the dummy extension must call this first.
func registerTestExtension() {
	registerExtensionOnce.Do(func() {
		extensions.Register(&extdef.Extension{
			Name: "doltgres_test",
			Control: extdef.Control{
				DefaultVersion: "1.0",
				Comment:        "test extension to ensure that emulated extensions behave properly",
				Relocatable:    true,
			},
			Types: []extdef.Type{
				{
					Name:       "dgtest_upper",
					Definition: pgtypes.NewBaseTypeDefinition(),
					Input:      "dgtest_upper_in",
					Output:     "dgtest_upper_out",
				},
			},
			Routines: []extdef.Routine{
				{
					Name:       "dgtest_upper_in",
					Symbol:     "dgtest_upper_in",
					Parameters: []extdef.Parameter{{Type: "cstring"}},
					Returns:    "dgtest_upper",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return strings.ToUpper(args[0].(string)), nil
					},
				},
				{
					Name:       "dgtest_upper_out",
					Symbol:     "dgtest_upper_out",
					Parameters: []extdef.Parameter{{Type: "dgtest_upper"}},
					Returns:    "cstring",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return args[0].(string), nil
					},
				},
				{
					Name:       "dgtest_upper_eq",
					Symbol:     "dgtest_upper_eq",
					Parameters: []extdef.Parameter{{Type: "dgtest_upper"}, {Type: "dgtest_upper"}},
					Returns:    "bool",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return args[0].(string) == args[1].(string), nil
					},
				},
				{
					Name:       "dgtest_upper_text",
					Symbol:     "dgtest_upper_text",
					Parameters: []extdef.Parameter{{Type: "dgtest_upper"}},
					Returns:    "text",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return args[0].(string), nil
					},
				},
				{
					Name:       "dgtest_upper_distance",
					Symbol:     "dgtest_upper_distance",
					Parameters: []extdef.Parameter{{Type: "dgtest_upper"}, {Type: "dgtest_upper"}},
					Returns:    "int4",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return int32(utils.Abs(len(args[0].(string)) - len(args[1].(string)))), nil
					},
				},
				{
					Name:       "dgtest_charcount_transition",
					Symbol:     "dgtest_charcount_transition",
					Parameters: []extdef.Parameter{{Type: "int4"}, {Type: "text"}},
					Returns:    "int4",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return args[0].(int32) + int32(len([]rune(args[1].(string)))), nil
					},
				},
				{
					Name:       "dgtest_charcount_combine",
					Symbol:     "dgtest_charcount_combine",
					Parameters: []extdef.Parameter{{Type: "int4"}, {Type: "int4"}},
					Returns:    "int4",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return args[0].(int32) + args[1].(int32), nil
					},
				},
				{
					Name:       "dgtest_charcount_final",
					Symbol:     "dgtest_charcount_final",
					Parameters: []extdef.Parameter{{Type: "int4"}},
					Returns:    "text",
					Strict:     true,
					Impl: func(ctx *sql.Context, args ...any) (any, error) {
						return strconv.Itoa(int(args[0].(int32))), nil
					},
				},
			},
			Operators: []extdef.Operator{
				{
					Symbol:     "=",
					Left:       "dgtest_upper",
					Right:      "dgtest_upper",
					Routine:    "dgtest_upper_eq",
					Commutator: "=",
					Hashes:     true,
					Merges:     true,
				},
				{
					Symbol:     "<->",
					Left:       "dgtest_upper",
					Right:      "dgtest_upper",
					Routine:    "dgtest_upper_distance",
					Commutator: "<->",
				},
			},
			Casts: []extdef.Cast{
				{
					Source:   "dgtest_upper",
					Target:   "text",
					Routine:  "dgtest_upper_text",
					CastType: casts.CastType_Assignment,
				},
			},
			Aggregates: []extdef.Aggregate{
				{
					Name:        "dgtest_charcount",
					Parameters:  []extdef.Parameter{{Type: "text"}},
					Returns:     "text",
					StateType:   "int4",
					Transition:  "dgtest_charcount_transition",
					Final:       "dgtest_charcount_final",
					Combine:     "dgtest_charcount_combine",
					InitCond:    "0",
					HasInitCond: true,
				},
			},
		})
	})
}
