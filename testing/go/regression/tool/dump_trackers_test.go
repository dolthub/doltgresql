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
	"os"
	"sort"
	"strings"
	"testing"
)

// TestDumpTrackers is a debugging tool rather than a unit test: it converts a binary .trackers file (produced by
// TestRegressionTests) into a human-readable text report, containing overall pass/fail counts, per-file pass rates
// sorted by failure count, and every failed or partially-successful query along with its expected and received
// errors. The text form is convenient for grepping and for bulk failure analysis (e.g. categorizing failures by
// error message to find common root causes).
//
// It only runs when the DUMP_TRACKERS environment variable is set, and asserts nothing. Usage:
//
//	DUMP_TRACKERS=/path/to/report.txt go test -run TestDumpTrackers
//
// By default it reads out/results.trackers (the file the regression run writes); set DUMP_TRACKERS_IN to dump a
// different trackers file, such as a saved baseline or one downloaded from a CI artifact.
func TestDumpTrackers(t *testing.T) {
	outPath, ok := os.LookupEnv("DUMP_TRACKERS")
	if !ok {
		t.Skip()
	}
	inPath := os.Getenv("DUMP_TRACKERS_IN")
	if inPath == "" {
		inPath = "out/results.trackers"
	}
	trackers, err := regressionFolder.ReadReplayTrackers(inPath)
	if err != nil {
		t.Fatal(err)
	}
	sb := &strings.Builder{}
	totalSuccess, totalFail, totalPartial := uint32(0), uint32(0), uint32(0)
	type fileStat struct {
		file          string
		success, fail uint32
	}
	stats := make([]fileStat, 0, len(trackers))
	for _, tr := range trackers {
		totalSuccess += tr.Success
		totalFail += tr.Failed
		totalPartial += tr.PartialSuccess
		stats = append(stats, fileStat{tr.File, tr.Success, tr.Failed})
	}
	total := totalSuccess + totalFail
	fmt.Fprintf(sb, "TOTAL: %d  SUCCESS: %d (%.2f%%)  FAIL: %d (%.2f%%)  PARTIAL: %d\n\n",
		total, totalSuccess, float64(totalSuccess)/float64(total)*100,
		totalFail, float64(totalFail)/float64(total)*100, totalPartial)
	sort.Slice(stats, func(i, j int) bool { return stats[i].fail > stats[j].fail })
	fmt.Fprintf(sb, "PER-FILE (sorted by failures):\n")
	for _, s := range stats {
		ft := s.success + s.fail
		if ft == 0 {
			continue
		}
		fmt.Fprintf(sb, "%-40s total=%-5d success=%-5d fail=%-5d (%.1f%% pass)\n",
			s.file, ft, s.success, s.fail, float64(s.success)/float64(ft)*100)
	}
	sb.WriteString("\n==================== FAILURE DETAILS ====================\n")
	for _, tr := range trackers {
		if len(tr.FailPartialItems) == 0 {
			continue
		}
		fmt.Fprintf(sb, "\n########## FILE: %s (fail=%d partial=%d) ##########\n", tr.File, tr.Failed, tr.PartialSuccess)
		for _, item := range tr.FailPartialItems {
			fmt.Fprintf(sb, "---\nQUERY: %s\n", item.Query)
			if item.ExpectedError != "" {
				fmt.Fprintf(sb, "EXPECTED ERROR: %s\n", item.ExpectedError)
			}
			if item.UnexpectedError != "" {
				fmt.Fprintf(sb, "RECEIVED ERROR: %s\n", item.UnexpectedError)
			}
			for _, p := range item.PartialSuccess {
				fmt.Fprintf(sb, "PARTIAL: %s\n", p)
			}
		}
	}
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		t.Fatal(err)
	}
	fmt.Println("wrote", outPath)
}
