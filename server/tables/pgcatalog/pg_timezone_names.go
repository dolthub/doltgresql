// Copyright 2024 Dolthub, Inc.
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

package pgcatalog

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgTimezoneNamesName is a constant to the pg_timezone_names name.
const PgTimezoneNamesName = "pg_timezone_names"

// InitPgTimezoneNames handles registration of the pg_timezone_names handler.
func InitPgTimezoneNames() {
	tables.AddHandler(PgCatalogName, PgTimezoneNamesName, PgTimezoneNamesHandler{})
}

// PgTimezoneNamesHandler is the handler for the pg_timezone_names table.
type PgTimezoneNamesHandler struct{}

var _ tables.Handler = PgTimezoneNamesHandler{}

// Name implements the interface tables.Handler.
func (p PgTimezoneNamesHandler) Name() string {
	return PgTimezoneNamesName
}

// RowIter implements the interface tables.Handler.
func (p PgTimezoneNamesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// The set of known timezone names cannot change while the process is running, so it is cached,
	// but the offset/abbreviation/DST flag reflect the current time (matching Postgres, which
	// reports the currently-observed offset for each zone).
	zones := loadTimezoneNames()
	now := time.Now()
	rows := make([]sql.Row, len(zones))
	for i, zone := range zones {
		localTime := now.In(zone.loc)
		abbrev, offsetSecs := localTime.Zone()
		rows[i] = sql.Row{
			zone.name, // name
			abbrev,    // abbrev
			duration.MakeDuration(int64(offsetSecs)*int64(time.Second), 0, 0), // utc_offset
			localTime.IsDST(), // is_dst
		}
	}
	return &pgTimezoneNamesRowIter{rows: rows}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgTimezoneNamesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTimezoneNamesSchema,
		PkOrdinals: nil,
	}
}

// pgTimezoneNamesSchema is the schema for pg_timezone_names.
var pgTimezoneNamesSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
	{Name: "abbrev", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
	{Name: "utc_offset", Type: pgtypes.Interval, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
	{Name: "is_dst", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
}

// pgTimezoneNamesRowIter is the sql.RowIter for the pg_timezone_names table.
type pgTimezoneNamesRowIter struct {
	rows []sql.Row
	idx  int
}

var _ sql.RowIter = (*pgTimezoneNamesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTimezoneNamesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.rows) {
		return nil, io.EOF
	}
	iter.idx++
	return iter.rows[iter.idx-1], nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTimezoneNamesRowIter) Close(ctx *sql.Context) error {
	return nil
}

// timezoneName is a single IANA timezone known to the system, along with its loaded location.
type timezoneName struct {
	name string
	loc  *time.Location
}

var (
	// timezoneNamesOnce guards the one-time loading of cachedTimezoneNames.
	timezoneNamesOnce sync.Once
	// cachedTimezoneNames is the cached, name-sorted list of timezones known to the system.
	cachedTimezoneNames []timezoneName
)

// zoneInfoDirs are the standard locations of the system timezone database on Unix systems. These
// are the same candidate sources that Go's time package searches.
var zoneInfoDirs = []string{
	"/usr/share/zoneinfo",
	"/usr/share/lib/zoneinfo",
	"/usr/lib/locale/TZ",
}

// loadTimezoneNames returns the cached, name-sorted list of IANA timezones known to the system.
// The list is loaded once per process, since the system timezone database cannot change while the
// process is running. If no timezone database is available, the list contains just UTC.
func loadTimezoneNames() []timezoneName {
	timezoneNamesOnce.Do(func() {
		var names []string
		for _, dir := range zoneInfoDirs {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				collectZoneNames(dir, "", &names)
				break
			}
		}
		zones := make([]timezoneName, 0, len(names)+1)
		seenUTC := false
		for _, name := range names {
			loc, err := time.LoadLocation(name)
			if err != nil {
				// Not a valid timezone file (e.g. leapseconds, SECURITY), so skip it
				continue
			}
			if name == "UTC" {
				seenUTC = true
			}
			zones = append(zones, timezoneName{name: name, loc: loc})
		}
		if !seenUTC {
			zones = append(zones, timezoneName{name: "UTC", loc: time.UTC})
		}
		sort.Slice(zones, func(i, j int) bool {
			return zones[i].name < zones[j].name
		})
		cachedTimezoneNames = zones
	})
	return cachedTimezoneNames
}

// collectZoneNames recursively walks the zoneinfo directory |root|, appending the relative paths of
// candidate timezone files under |prefix| to |names|. Non-timezone entries are filtered using the
// same rules as Go's time package and Postgres: names containing '.' (zone.tab, tzdata.zi, etc.),
// the posixrules and localtime files, and the top-level right/, posix/, and Factory entries.
func collectZoneNames(root string, prefix string, names *[]string) {
	entries, err := os.ReadDir(filepath.Join(root, prefix))
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, ".") || name == "posixrules" || name == "localtime" {
			continue
		}
		if prefix == "" && (name == "right" || name == "posix" || name == "Factory") {
			continue
		}
		fullName := name
		if prefix != "" {
			fullName = prefix + "/" + name
		}
		// os.Stat (rather than entry.IsDir) so that symlinked directories are followed
		info, err := os.Stat(filepath.Join(root, fullName))
		if err != nil {
			continue
		}
		if info.IsDir() {
			collectZoneNames(root, fullName, names)
		} else {
			*names = append(*names, fullName)
		}
	}
}
