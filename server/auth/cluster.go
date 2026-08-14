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

package auth

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/cluster"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
)

// clusterReplicator, when non-nil, takes over persistence of the auth database: local writes flow through it so
// that they are also replicated to cluster standbys.
var clusterReplicator cluster.ReplicatingAuthPersister

// SetClusterReplicator routes all subsequent auth database persistence through |replicator|, which persists the
// serialized state locally and replicates it to cluster standbys. Called during server startup when cluster
// replication is configured.
func SetClusterReplicator(replicator cluster.ReplicatingAuthPersister) {
	clusterReplicator = replicator
}

// WaitForReplication blocks until the replication acks accumulated in |rsc| complete, subject to the
// dolt_cluster_ack_writes_timeout_secs system variable. It is a no-op when |rsc| holds no waiters (e.g. when
// cluster replication is not enabled). Callers must not hold the auth database lock.
func WaitForReplication(ctx *sql.Context, rsc doltdb.ReplicationStatusController) {
	dsess.WaitForReplicationController(ctx, rsc)
}

// ReadSerializedDatabase returns the serialized auth database as persisted to the auth database file, or nil when
// running with the pure in-memory implementation. The file is written on every auth change, so its contents match
// the live state. Unlike serializing the live state, reading the file takes no locks, which matters because
// cluster replication calls this while holding its own locks (see PersistChanges, which is called with the auth
// write lock held and takes replication locks — the opposite order).
func ReadSerializedDatabase() ([]byte, error) {
	if fileSystem == nil {
		return nil, nil
	}
	return fileSystem.ReadFile(authFileName)
}

// WriteSerializedDatabase writes an already-serialized auth database payload to the auth database file. It does
// not touch the in-memory state; use OverwriteDatabase to replace both.
func WriteSerializedDatabase(data []byte) error {
	if fileSystem != nil {
		return fileSystem.WriteFile(authFileName, data, 0644)
	}
	return nil
}

// OverwriteDatabase replaces the entire contents of the auth database, both in memory and on disk, with the
// deserialized contents of |data|. This is used on cluster standbys to apply the auth state replicated from the
// primary. The previous state is retained if deserialization fails.
func OverwriteDatabase(data []byte) error {
	var err error
	LockWrite(func() {
		fresh := newEmptyDatabase()
		if err = fresh.deserialize(data); err != nil {
			return
		}
		globalDatabase = fresh
		err = WriteSerializedDatabase(data)
	})
	return err
}
