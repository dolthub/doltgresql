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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/cluster"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/auth"
)

// authDbPersister adapts Doltgres's auth database (auth.db) to dolt's cluster.AuthDbPersister, making it the
// local source and sink for the serialized auth payload that cluster replication carries.
type authDbPersister struct{}

var _ cluster.AuthDbPersister = authDbPersister{}

// Persist implements cluster.AuthDbPersister. It receives the payload that was just serialized from the live
// auth database and writes it to the auth database file.
func (authDbPersister) Persist(_ *sql.Context, data []byte) error {
	return auth.WriteSerializedDatabase(data)
}

// LoadData implements cluster.AuthDbPersister. It is called when this server becomes a cluster primary, to seed
// replication with the current auth state. It reads the auth database file rather than serializing the live
// state: the file is written on every auth change, and reading it takes no locks, which matters because the
// replication layer calls this while holding locks that the write path acquires after the auth lock.
func (authDbPersister) LoadData(context.Context) ([]byte, error) {
	return auth.ReadSerializedDatabase()
}

// authDbPersistence implements cluster.AuthPersistence: it applies an auth payload replicated from the cluster
// primary to this standby, replacing the in-memory state and the auth database file.
type authDbPersistence struct{}

var _ cluster.AuthPersistence = authDbPersistence{}

// SaveData implements cluster.AuthPersistence.
func (authDbPersistence) SaveData(_ *sql.Context, contents []byte) error {
	return auth.OverwriteDatabase(contents)
}

// hookAuthReplication takes over cluster auth replication from the default users-and-grants implementation
// installed by dolt: Doltgres stores auth data in auth.db rather than in the engine's mysql database. Local auth
// writes are routed through the controller's replicating persister so that they reach the standbys, and payloads
// replicated from a primary are applied to auth.db. The initial LoadData seeds replication with the current auth
// state, so that a primary pushes its full auth database to the standbys at startup.
func hookAuthReplication(ctx context.Context, controller *cluster.Controller) error {
	if controller == nil {
		return nil
	}
	replicator := controller.HookAuthPersister(authDbPersister{}, authDbPersistence{})
	auth.SetClusterReplicator(replicator)
	_, err := replicator.LoadData(ctx)
	return err
}
