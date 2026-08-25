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
	"github.com/dolthub/doltgresql/core/integrity"

	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
)

// databaseReport holds the scan (and optionally repair) results for one database.
type databaseReport struct {
	Name     string
	Branches []*branchReport
	// Repair is non-nil when the tool ran in repair mode.
	Repair *repairSummary
	// PostRepair holds the re-scan of every branch after repair, when the tool ran in repair mode.
	PostRepair []*branchReport
}

// branchReport holds the scan results for every table at the head of one branch.
type branchReport struct {
	Branch string
	Tables []*tableReport
}

// tableReport holds the scan results for one table.
type tableReport struct {
	Schema string
	Table  string
	// AdaptiveValueCols and AdaptiveKeyCols name the adaptive-encoded columns in the table's schema.
	AdaptiveValueCols []string
	AdaptiveKeyCols   []string
	// Impacted is true when the table's value tuples contain adaptive-encoded fields, making it
	// subject to the value_address_offsets corruption.
	Impacted bool
	// KeyImpacted is true when the table's key tuples contain adaptive-encoded fields. Out-of-band
	// key values are structurally untracked (there is no key_address_offsets field), so they are
	// reported but not repairable by rewriting nodes.
	KeyImpacted bool
	// Stats is nil when the table is not impacted (it is not scanned).
	Stats *integrity.Stats
	Err   string
}

var reportFuncs = template.FuncMap{
	"pct": func(part, whole uint64) string {
		if whole == 0 {
			return "0.00%"
		}
		return fmt.Sprintf("%.2f%%", float64(part)/float64(whole)*100.0)
	},
	"join": func(ss []string) string {
		return strings.Join(ss, ", ")
	},
}

var reportTemplate = template.Must(template.New("report").Funcs(reportFuncs).Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Adaptive value_address_offsets corruption report</title>
<style>
body { font-family: sans-serif; margin: 2em; }
h1 { font-size: 1.4em; }
h2 { font-size: 1.2em; margin-top: 1.5em; }
h3 { font-size: 1.0em; margin-top: 1em; }
table { border-collapse: collapse; margin: 0.5em 0 1.5em 0; }
th, td { border: 1px solid #ccc; padding: 4px 10px; text-align: right; font-size: 0.9em; }
th { background: #f0f0f0; }
td.name { text-align: left; font-family: monospace; }
tr.corrupt td { background: #fff0f0; }
tr.clean td { background: #f4fff4; }
tr.skipped td { color: #999; }
.summary { background: #f8f8f8; border: 1px solid #ddd; padding: 0.75em 1em; margin: 1em 0; }
.warn { color: #a00; }
</style>
</head>
<body>
<h1>Adaptive value_address_offsets corruption report</h1>
<p>Generated {{.Generated}} · mode: {{.Mode}}</p>
{{range .Databases}}
<h2>Database: {{.Name}}</h2>
{{if .Repair}}
<div class="summary">
<b>Repair summary:</b>
commits examined: {{.Repair.CommitsExamined}},
commits rewritten: {{.Repair.CommitsRewritten}},
branches updated: {{.Repair.BranchesUpdated}},
tags updated: {{.Repair.TagsUpdated}},
working sets repaired: {{.Repair.WorkingSetsFixed}},
leaf chunks rewritten: {{.Repair.LeafChunksRewritten}},
internal chunks rewritten: {{.Repair.InternalChunksRewritten}}
</div>
{{end}}
{{template "branches" .Branches}}
{{if .PostRepair}}
<h3>Post-repair verification scan</h3>
{{template "branches" .PostRepair}}
{{end}}
{{end}}
</body>
</html>

{{define "branches"}}
{{range .}}
<h3>Branch: {{.Branch}}</h3>
<table>
<tr>
<th>Schema</th><th>Table</th><th>Adaptive columns</th>
<th>Rows</th><th>Corrupt rows</th><th>% rows</th>
<th>Adaptive values</th><th>Out-of-band</th><th>Corrupt values</th><th>% values</th>
<th>Chunks</th><th>Corrupt chunks</th>
<th>Key out-of-band</th><th>Missing chunks</th>
</tr>
{{range .Tables}}
{{if .Err}}
<tr class="corrupt"><td class="name">{{.Schema}}</td><td class="name">{{.Table}}</td><td colspan="12" class="warn">error: {{.Err}}</td></tr>
{{else if not .Stats}}
<tr class="skipped"><td class="name">{{.Schema}}</td><td class="name">{{.Table}}</td><td class="name">(schema not impacted)</td><td colspan="11"></td></tr>
{{else}}
<tr class="{{if .Stats.CorruptValues}}corrupt{{else}}clean{{end}}">
<td class="name">{{.Schema}}</td>
<td class="name">{{.Table}}</td>
<td class="name">{{join .AdaptiveValueCols}}{{if .AdaptiveKeyCols}} [key: {{join .AdaptiveKeyCols}}]{{end}}</td>
<td>{{.Stats.Rows}}</td>
<td>{{.Stats.CorruptRows}}</td>
<td>{{pct .Stats.CorruptRows .Stats.Rows}}</td>
<td>{{.Stats.AdaptiveValues}}</td>
<td>{{.Stats.OutOfBandValues}}</td>
<td>{{.Stats.CorruptValues}}</td>
<td>{{pct .Stats.CorruptValues .Stats.AdaptiveValues}}</td>
<td>{{.Stats.Chunks}}</td>
<td>{{.Stats.CorruptChunks}}</td>
<td>{{.Stats.KeyOutOfBandValues}}</td>
<td>{{if .Stats.MissingChunks}}<span class="warn">{{.Stats.MissingChunks}}</span>{{else}}0{{end}}</td>
</tr>
{{end}}
{{end}}
</table>
{{end}}
{{end}}
`))

// writeHTMLReport renders the report for all databases to |w|.
func writeHTMLReport(w io.Writer, mode string, dbs []*databaseReport) error {
	return reportTemplate.Execute(w, struct {
		Generated string
		Mode      string
		Databases []*databaseReport
	}{
		Generated: time.Now().Format(time.RFC1123),
		Mode:      mode,
		Databases: dbs,
	})
}
