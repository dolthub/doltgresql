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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestReportOmitsCleanTables verifies that the HTML report only mentions tables with missing
// values: scanned-but-clean tables are counted in the branch summary but not listed.
func TestReportOmitsCleanTables(t *testing.T) {
	report := &Report{
		Mode:        "report",
		DataDir:     "/data",
		GeneratedAt: time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC),
		Databases: []*DatabaseReport{{
			Database: "db1",
			Branches: []*BranchReport{
				{
					Branch:         "main",
					CommitsScanned: 1,
					SkippedTables:  3,
					Tables: []*TableStats{
						{Table: "public.clean_table", AdaptiveColumns: []string{"big"}, RowsScanned: 10, AdaptiveValues: 10},
						{Table: "public.broken_table", AdaptiveColumns: []string{"big"}, RowsScanned: 10, RowsWithMissing: 1,
							AdaptiveValues: 10, OutOfBandValues: 4, MissingValues: 1,
							MissingByColumn: map[string]uint64{"big": 1}},
					},
				},
				{
					Branch:         "empty",
					CommitsScanned: 1,
					Tables: []*TableStats{
						{Table: "public.clean_table", AdaptiveColumns: []string{"big"}, RowsScanned: 10, AdaptiveValues: 10},
					},
				},
			},
		}},
	}

	var sb strings.Builder
	require.NoError(t, report.WriteHTML(&sb))
	html := sb.String()

	require.Contains(t, html, "public.broken_table")
	require.NotContains(t, html, "public.clean_table", "clean tables must be omitted from the report")
	require.Contains(t, html, "No missing adaptive values on this branch.", "branches with no impacted tables get a placeholder")
	require.Contains(t, html, "1 adaptive table clean", "clean tables are still counted in the branch summary")
}
