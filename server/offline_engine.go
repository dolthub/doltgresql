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

package server

import (
	"context"

	"github.com/dolthub/dolt/go/cmd/dolt/commands/engine"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/auth"
)

// NewOfflineSqlEngine constructs a SqlEngine over the databases in |mrEnv| for offline administrative
// use: no server, no replication, and no background garbage collection. Sessions created from it are
// the same kind a running server would create for a query, which storage-level code requires in some
// cases: for example, deserializing a table schema that uses a user-defined data type resolves the type
// through the session's Doltgres state, and panics on a bare session. The caller must Close the
// returned engine. initialization.Initialize must have been called first.
//
// This is useful for certain offline, non-server admin tasks where starting a server is either undesirable
// or impossible.
func NewOfflineSqlEngine(ctx context.Context, mrEnv *env.MultiRepoEnv) (*engine.SqlEngine, error) {
	user, _ := auth.GetSuperUserAndPassword()
	return engine.NewSqlEngine(ctx, mrEnv, &engine.SqlEngineConfig{
		ServerUser:      user,
		Autocommit:      true,
		ProviderFactory: DoltgresProviderFactory{},
	})
}

// NewOfflineSessionContext returns a *sql.Context backed by a real session from |se|, running as the
// Doltgres superuser, along with a cleanup function that ends the session. Callers examining a specific
// database must also set it with ctx.SetCurrentDatabase, which user-defined type resolution depends on.
func NewOfflineSessionContext(ctx context.Context, se *engine.SqlEngine) (*sql.Context, func(), error) {
	sctx, err := se.NewDefaultContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	// The session must run as the doltgres superuser: unlike Dolt, doltgres has no "root" user, so
	// the default client would be rejected.
	user, _ := auth.GetSuperUserAndPassword()
	sctx.Session.SetClient(sql.Client{User: user, Address: "localhost", Capabilities: 0})
	sql.SessionCommandBegin(sctx.Session)
	cleanup := func() {
		sql.SessionCommandEnd(sctx.Session)
		sql.SessionEnd(sctx.Session)
	}
	return sctx, cleanup, nil
}
