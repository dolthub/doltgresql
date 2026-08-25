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

package main

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/dbfactory"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/store/chunks"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/server/initialization"
	doltgresservercfg "github.com/dolthub/doltgresql/servercfg"
)

// database bundles the storage-layer handles for a single database found in the data directory.
type database struct {
	name string
	dEnv *env.DoltEnv
	ddb  *doltdb.DoltDB
	cs   chunks.ChunkStore
	ns   tree.NodeStore
}

// openDatabases initializes the doltgres runtime (type serialization etc.) and opens every database found in |dir|.
// |dir| may either be a database directory itself (contains .dolt) or a data directory containing databases.
func openDatabases(ctx context.Context, dir string) ([]*database, error) {
	initialization.Initialize(nil, doltgresservercfg.DefaultServerConfig())

	fs, err := filesys.LocalFilesysWithWorkingDir(dir)
	if err != nil {
		return nil, err
	}

	dEnv := env.LoadWithoutDB(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, "doltgres-admin")
	// The singleton database cache returns closed stores if the same database is opened twice in one
	// process (e.g. in tests that reopen after a server shuts down), so disable it.
	dEnv.DBLoadParams = map[string]interface{}{dbfactory.DisableSingletonCacheParam: struct{}{}}

	mrEnv, err := env.MultiEnvForDirectory(ctx, fs, dEnv)
	if err != nil {
		return nil, err
	}

	var dbs []*database
	err = mrEnv.Iter(func(name string, de *env.DoltEnv) (bool, error) {
		if de.DBLoadError != nil {
			return true, errors.Wrapf(de.DBLoadError, "failed to load database %s", name)
		}
		ddb := de.DoltDB(ctx)
		if ddb == nil {
			return false, nil
		}
		cs := datas.ChunkStoreFromDatabase(doltdb.ExposeDatabaseFromDoltDB(ddb))
		dbs = append(dbs, &database{
			name: name,
			dEnv: de,
			ddb:  ddb,
			cs:   cs,
			ns:   ddb.NodeStore(),
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	if len(dbs) == 0 {
		return nil, errors.Errorf("no databases found in %s", dir)
	}
	return dbs, nil
}

// close closes the underlying dolt databases.
func closeDatabases(dbs []*database) {
	for _, db := range dbs {
		_ = db.ddb.Close()
	}
}
