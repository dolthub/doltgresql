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
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"
)

// WriteHTML renders the report as a standalone HTML page.
func (r *Report) WriteHTML(w io.Writer) error {
	return reportTemplate.Execute(w, r)
}

var templateFuncs = template.FuncMap{
	"pct": func(f float64) string {
		return fmt.Sprintf("%.2f%%", f)
	},
	"columnBreakdown": func(m map[string]uint64) string {
		if len(m) == 0 {
			return ""
		}
		cols := make([]string, 0, len(m))
		for col := range m {
			cols = append(cols, col)
		}
		sort.Strings(cols)
		parts := make([]string, len(cols))
		for i, col := range cols {
			parts[i] = fmt.Sprintf("%s: %d", col, m[col])
		}
		return strings.Join(parts, ", ")
	},
	"join": strings.Join,
}

var reportTemplate = template.Must(template.New("report").Funcs(templateFuncs).Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Adaptive Encoding Corruption {{if eq .Mode "repair"}}Repair{{else}}Report{{end}}</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
         margin: 2em; color: #1f2328; }
  h1 { font-size: 1.5em; }
  h2 { font-size: 1.25em; margin-top: 1.5em; border-bottom: 1px solid #d1d9e0; padding-bottom: 0.3em; }
  h3 { font-size: 1.05em; margin-top: 1.2em; }
  .meta { color: #59636e; font-size: 0.9em; }
  table { border-collapse: collapse; margin: 0.75em 0 1.5em 0; font-size: 0.9em; }
  th, td { border: 1px solid #d1d9e0; padding: 0.4em 0.75em; text-align: right; }
  th { background: #f6f8fa; }
  td.name, th.name { text-align: left; }
  tr.impacted td { background: #fff8f8; }
  td.bad { color: #d1242f; font-weight: 600; }
  td.ok { color: #1a7f37; }
  .error { color: #d1242f; }
  .none { color: #59636e; font-style: italic; }
</style>
</head>
<body>
<h1>Adaptive Encoding Corruption {{if eq .Mode "repair"}}Repair{{else}}Report{{end}}</h1>
<p class="meta">Mode: {{.Mode}} &middot; Data directory: <code>{{.DataDir}}</code> &middot; Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</p>
{{range .Databases}}
<h2>Database: {{.Database}}</h2>
{{range .Errors}}<p class="error">Error: {{.}}</p>{{end}}
{{if not .Branches}}<p class="none">No branches found.</p>{{end}}
{{range .Branches}}
<h3>Branch: {{.Branch}} <span class="meta">({{.CommitsScanned}} commit{{if ne .CommitsScanned 1}}s{{end}} scanned; {{.CleanTableCount}} adaptive table{{if ne .CleanTableCount 1}}s{{end}} clean; {{.SkippedTables}} table{{if ne .SkippedTables 1}}s{{end}} without adaptive columns skipped)</span></h3>
{{if not .ImpactedTables}}
<p class="none">No missing adaptive values on this branch.</p>
{{else}}
<table>
  <tr>
    <th class="name">Table</th>
    <th class="name">Adaptive Columns</th>
    <th>Rows Scanned</th>
    <th>Rows w/ Missing</th>
    <th>% Rows Missing</th>
    <th>Adaptive Values</th>
    <th>Out-of-band</th>
    <th>Missing</th>
    <th>Missing (Key Cols)</th>
    <th>% Values Missing</th>
    {{if eq $.Mode "repair"}}<th>Repaired (NULLed)</th>{{end}}
  </tr>
  {{range .ImpactedTables}}
  <tr class="impacted">
    <td class="name">{{.Table}}</td>
    <td class="name">{{join .AdaptiveColumns ", "}}{{if .AdaptiveKeyColumns}}{{if .AdaptiveColumns}} {{end}}(key: {{join .AdaptiveKeyColumns ", "}}){{end}}</td>
    <td>{{.RowsScanned}}</td>
    <td>{{.RowsWithMissing}}</td>
    <td>{{pct .PctRowsMissing}}</td>
    <td>{{.AdaptiveValues}}</td>
    <td>{{.OutOfBandValues}}</td>
    <td{{if .MissingValues}} class="bad"{{else}} class="ok"{{end}}>{{.MissingValues}}</td>
    <td{{if .MissingKeyValues}} class="bad"{{end}}>{{.MissingKeyValues}}</td>
    <td>{{pct .PctValuesMissing}}</td>
    {{if eq $.Mode "repair"}}<td>{{.RepairedValues}}</td>{{end}}
  </tr>
  {{if .MissingByColumn}}
  <tr><td class="name" colspan="{{if eq $.Mode "repair"}}11{{else}}10{{end}}">&nbsp;&nbsp;&#8627; missing by column &mdash; {{columnBreakdown .MissingByColumn}}</td></tr>
  {{end}}
  {{end}}
</table>
{{end}}
{{end}}
{{end}}
</body>
</html>
`))
