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
var _ cluster.ReplicaAuthPersister = authDbPersister{}

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

// SaveData implements cluster.ReplicaAuthPersister.
func (authDbPersister) SaveData(_ *sql.Context, contents []byte) error {
	return auth.OverwriteDatabase(contents)
}

// configureAuthReplication installs doltgres's auth.db persister for cluster replication
func configureAuthReplication(ctx context.Context, controller *cluster.Controller) error {
	if controller == nil {
		return nil
	}

	var persister authDbPersister

	replicator := controller.SetAuthReplicator(persister, persister)
	auth.SetClusterReplicator(replicator)
	_, err := replicator.LoadData(ctx)
	return err
}
