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

package integrity

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/go-mysql-server/sql"
)

// CorruptionError describes the first corrupted table found by CheckDatabase, with instructions for
// repairing the database.
type CorruptionError struct {
	Database string
	Branch   string
	// Commit is the hash of the corrupt commit, or empty if the corruption was found in the branch's
	// uncommitted working set.
	Commit string
	Table  string
	Stats  *Stats
}

var _ error = (*CorruptionError)(nil)

func (e *CorruptionError) Error() string {
	location := fmt.Sprintf("commit %s (reachable from branch %s)", e.Commit, e.Branch)
	if e.Commit == "" {
		location = fmt.Sprintf("the working set of branch %s", e.Branch)
	}
	return fmt.Sprintf(`database %q failed the startup integrity check: table %s at %s has %d adaptive-encoded values (in %d rows) that are missing from the storage layer's chunk reference index (value_address_offsets). This corruption was written by an earlier release; pushing, cloning, or garbage collecting this database will lose the affected values.

To repair this database:
  1. Stop the server and MAKE A BACKUP COPY of the database directory before doing anything else.
  2. Do not run garbage collection before repairing.
  3. Build the repair tool from the DoltgreSQL repository: go build -o doltgres-admin ./cmd/admin
  4. Run: doltgres-admin report -dir <data-dir> to see the full extent of the corruption, then
     doltgres-admin repair -dir <data-dir> to repair it.

To start the server without this check (unsafe until the database is repaired), set behavior.skip_startup_integrity_check to true in config.yaml.`,
		e.Database, e.Table, location, e.Stats.CorruptValues, e.Stats.CorruptRows)
}

// CheckDatabase scans every table in every commit reachable from every branch of the database, as well
// as each branch's working set, for adaptive-encoded out-of-band values that are missing from their
// node's value_address_offsets field. It returns a *CorruptionError describing the first corrupted
// table found, or nil if the database is healthy. Scan results are cached per chunk, so shared history
// and structurally identical tables cost nothing to re-examine.
//
// This check must run before any garbage collection: GC treats the missing references as unreachable
// and deletes the affected values.
func CheckDatabase(sctx *sql.Context, dbName string, ddb *doltdb.DoltDB) error {
	cs := datas.ChunkStoreFromDatabase(doltdb.ExposeDatabaseFromDoltDB(ddb))
	sc := NewScanner(cs)

	branches, err := ddb.GetBranches(sctx)
	if err != nil {
		return err
	}

	visited := make(map[hash.Hash]struct{})
	for _, branchRef := range branches {
		head, err := ddb.ResolveCommitRef(sctx, branchRef)
		if err != nil {
			return errors.Wrapf(err, "failed to resolve branch %s", branchRef.String())
		}

		// Walk every commit reachable from this branch head.
		stack := []*doltdb.Commit{head}
		for len(stack) > 0 {
			cm := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			h, err := cm.HashOf()
			if err != nil {
				return err
			}
			if _, ok := visited[h]; ok {
				continue
			}
			visited[h] = struct{}{}

			root, err := cm.GetRootValue(sctx)
			if err != nil {
				return err
			}
			err = checkRoot(sctx, sc, dbName, branchRef.GetPath(), h.String(), root, ddb)
			if err != nil {
				return err
			}

			optParents, err := ddb.ResolveAllParents(sctx, cm)
			if err != nil {
				return err
			}
			for _, opt := range optParents {
				parent, ok := opt.ToCommit()
				if !ok {
					// Ghost commits (shallow clones) have no content to scan.
					continue
				}
				stack = append(stack, parent)
			}
		}

		// Check the branch's working set roots, which may contain uncommitted corrupt data.
		wsRef, err := ref.WorkingSetRefForHead(branchRef)
		if err != nil {
			return err
		}
		ws, err := ddb.ResolveWorkingSet(sctx, wsRef)
		if err != nil {
			// Not all branches have working sets (e.g. branches never checked out).
			continue
		}
		for _, wsRoot := range []doltdb.RootValue{ws.WorkingRoot(), ws.StagedRoot()} {
			if wsRoot == nil {
				continue
			}
			err = checkRoot(sctx, sc, dbName, branchRef.GetPath(), "", wsRoot, ddb)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// checkRoot scans every impacted table in the given root value, returning a *CorruptionError for the
// first corrupted table found.
func checkRoot(sctx *sql.Context, sc *Scanner, dbName, branch, commit string, root doltdb.RootValue, ddb *doltdb.DoltDB) error {
	tables, err := TablesForRoot(sctx, root, ddb.NodeStore())
	if err != nil {
		return err
	}
	for _, ti := range tables {
		if !ti.Impacted() {
			continue
		}
		stats, err := sc.ScanTable(sctx, ti)
		if err != nil {
			return errors.Wrapf(err, "failed to scan table %s", ti.Name.String())
		}
		if stats.CorruptValues > 0 {
			return &CorruptionError{
				Database: dbName,
				Branch:   branch,
				Commit:   commit,
				Table:    ti.Name.String(),
				Stats:    stats,
			}
		}
	}
	return nil
}
