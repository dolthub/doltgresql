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
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestRoleSettingsResetDefaults(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	for _, statement := range []string{
		"ALTER ROLE postgres SET work_mem = '16MB'",
		"ALTER ROLE postgres SET app.tenant = 'acme'",
	} {
		_, err := conn.Default.Exec(ctx, statement)
		require.NoError(t, err)
	}
	c, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/postgres", conn.Default.Config().Port))
	require.NoError(t, err)
	defer c.Close(ctx)
	for _, tc := range []struct{ name, expected string }{
		{"work_mem", "16384"}, {"app.tenant", "acme"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			require.NoError(t, c.QueryRow(ctx, "SELECT current_setting($1)", tc.name).Scan(&got))
			require.Equal(t, tc.expected, got)
			_, err := c.Exec(ctx, "RESET "+tc.name)
			require.NoError(t, err)
			require.NoError(t, c.QueryRow(ctx, "SELECT current_setting($1)", tc.name).Scan(&got))
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestRoleSettingsClientDateStyleOverride(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	_, err := conn.Default.Exec(ctx, "ALTER ROLE postgres SET DateStyle = 'SQL, DMY'")
	require.NoError(t, err)
	cfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/postgres", conn.Default.Config().Port))
	require.NoError(t, err)
	cfg.RuntimeParams["DateStyle"] = "ISO, MDY"
	c, err := pgx.ConnectConfig(ctx, cfg)
	require.NoError(t, err)
	defer c.Close(ctx)
	var got string
	require.NoError(t, c.QueryRow(ctx, "SELECT current_setting('DateStyle')").Scan(&got))
	require.Equal(t, "ISO, MDY", got)
}
