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

// runStartupIntegrityCheck scans every database in the data directory for errors in data integrity that should
// prevent the server from starting. It returns an error if it finds any.
func runStartupIntegrityCheck(ctx context.Context, mrEnv *env.MultiRepoEnv) error {
	start := time.Now()
	// Doltgres extended type deserialization requires a *sql.Context.
	// TODO: this is broken, we need to get an actual Doltgres session from the session manager
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
