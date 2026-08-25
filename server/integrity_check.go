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
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/integrity"
)

// runStartupIntegrityCheck scans every database in the data directory for adaptive-encoded out-of-band
// values that are missing from their tree node's chunk reference index (value_address_offsets), a form
// of corruption written by earlier releases. It returns an error describing the first corruption found,
// which prevents the server from starting.
//
// This check must run before any garbage collection process begins (auto-GC runs as part of the
// sql-server services): GC treats the missing references as unreachable and permanently deletes the
// affected values. It can be skipped for faster startup with behavior.skip_startup_integrity_check.
func runStartupIntegrityCheck(ctx context.Context, mrEnv *env.MultiRepoEnv) error {
	start := time.Now()
	// Doltgres extended type deserialization requires a *sql.Context.
	sctx := sql.NewContext(ctx)

	err := mrEnv.Iter(func(name string, de *env.DoltEnv) (bool, error) {
		ddb := de.DoltDB(ctx)
		if ddb == nil {
			return false, nil
		}
		if err := integrity.CheckDatabase(sctx, name, ddb); err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	logrus.Infof("startup integrity check passed in %s", time.Since(start))
	return nil
}
