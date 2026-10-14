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
	"encoding/binary"
	"strconv"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
)

func TestRegroleTextConversions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "quoted role output roundtrip",
			SetUpScript: []string{`CREATE ROLE "MixedCase"`, `CREATE ROLE mixedcase`},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT '"MixedCase"'::regrole::text`, Expected: []sql.Row{{`"MixedCase"`}}},
				{Query: `SELECT '"MixedCase"'::regrole::text::regrole = '"MixedCase"'::regrole`, Expected: []sql.Row{{"t"}}},
			},
		},
		{
			Name: "zero OID sentinel",
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT '-'::regrole::oid`, Expected: []sql.Row{{0}}},
				{Query: `SELECT 0::regrole::text`, Expected: []sql.Row{{"-"}}},
			},
		},
		{
			Name:        "quoted char role",
			SetUpScript: []string{`CREATE ROLE "char"`},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT '"char"'::regrole::oid = (SELECT oid FROM pg_roles WHERE rolname = 'char')`, Expected: []sql.Row{{"t"}}},
			},
		},
	})
}

func TestRegroleBinaryUncachedOID(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	// Select an OID that has not already been registered by another test.
	uncachedOID := uint32(424242)
	for id.Cache().ToInternal(uncachedOID).IsValid() {
		uncachedOID++
	}
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uncachedOID)
	result := conn.Default.PgConn().ExecParams(ctx, "SELECT $1::oid::text", [][]byte{data}, []uint32{uint32(oid.T_regrole)}, []int16{1}, []int16{0}).Read()
	require.NoError(t, result.Err)
	require.Equal(t, [][][]byte{{[]byte(strconv.FormatUint(uint64(uncachedOID), 10))}}, result.Rows)
}
