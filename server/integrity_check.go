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
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/integrity"
)

// runStartupIntegrityCheck scans every database in the data directory for errors in data integrity that should
// prevent the server from starting. It returns an error if it finds any.
func runStartupIntegrityCheck(ctx context.Context, mrEnv *env.MultiRepoEnv) error {
	start := time.Now()

	// Storage-level scans deserialize table schemas, which requires a real session of the kind the
	// server builds for a query (e.g. to resolve user-defined data types), so build a short-lived
	// engine over the same databases. It runs no background services, so no garbage collection can
	// begin before the check completes.
	se, err := NewOfflineSqlEngine(ctx, mrEnv)
	if err != nil {
		return err
	}
	defer se.Close()
	sctx, cleanup, err := NewOfflineSessionContext(ctx, se)
	if err != nil {
		return err
	}
	defer cleanup()

	var checked, skipped int
	err = mrEnv.Iter(func(name string, de *env.DoltEnv) (bool, error) {
		ddb := de.DoltDB(ctx)
		if ddb == nil {
			return false, nil
		}

		// Databases that already passed this version of the check carry a sentinel file and are
		// not checked again.
		doltDir := de.GetDoltDir()
		if integrity.HasValidSentinel(de.FS, doltDir) {
			logrus.Debugf("skipping integrity check for database %s: previously passed", name)
			skipped++
			return false, nil
		}

		// User-defined type resolution consults the session state of the current database.
		sctx.SetCurrentDatabase(name)
		if err := integrity.CheckDatabase(sctx, name, ddb); err != nil {
			return true, err
		}
		checked++

		// Record the result so future startups skip this database. Best effort: a database that
		// can't record it (e.g. a read-only filesystem) is just checked again next time.
		if err := integrity.WriteSentinel(de.FS, doltDir); err != nil {
			logrus.Warnf("failed to record integrity check result for database %s: %s", name, err.Error())
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	logrus.Infof("startup integrity check passed in %s (%d checked, %d previously passed)",
		time.Since(start), checked, skipped)
	return nil
}
