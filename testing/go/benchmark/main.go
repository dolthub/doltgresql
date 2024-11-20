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

package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// AllowedVariance represents the amount that must change before a result is noteworthy. The number represents a whole
// percentage, so "10" equals "10%".
const AllowedVariance = 10

// main analyzes two separate runs of scripts/quick_sysbench.sh, and creates a table comparing the differences. This
// table is intended for display in a GitHub PR.
func main() {
	prBenchmark, err := benchmarkFolder.ReadFile("results1.log")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	mainBenchmark, err := benchmarkFolder.ReadFile("results2.log")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	prSections := SectionResults(string(prBenchmark))
	mainSections := SectionResults(string(mainBenchmark))
	sb := strings.Builder{}
	sb.WriteString("|   | Main | PR |   |\n")
	sb.WriteString("| --- | --- | --- | --- |\n")
	for _, kv := range GetMapKVsSorted(mainSections) {
		mainSection := kv.Value
		prSection, ok := prSections[mainSection.Test]
		if !ok {
			sb.WriteString(fmt.Sprintf("| %s | %.2f/s | ${\\color{red}DNF}$ |   |\n",
				mainSection.Test, mainSection.Time))
			continue
		}
		percentChange := math.Floor(((prSection.Time/mainSection.Time)-1.0)*1000.0) / 10.0
		if percentChange > AllowedVariance { // Greatly positive
			sb.WriteString(fmt.Sprintf("| %s | %.2f/s | ${\\color{lightgreen}%.2f/s}$ | ${\\color{lightgreen}+%.1f\\\\%%}$ |\n",
				mainSection.Test, mainSection.Time, prSection.Time, percentChange))
		} else if percentChange < -AllowedVariance { // Greatly negative
			sb.WriteString(fmt.Sprintf("| %s | %.2f/s | ${\\color{red}%.2f/s}$ | ${\\color{red}%.1f\\\\%%}$ |\n",
				mainSection.Test, mainSection.Time, prSection.Time, percentChange))
		} else if percentChange > 0 { // Positive
			sb.WriteString(fmt.Sprintf("| %s | %.2f/s | %.2f/s | +%.1f%% |\n",
				mainSection.Test, mainSection.Time, prSection.Time, percentChange))
		} else if percentChange < 0 { // Negative
			sb.WriteString(fmt.Sprintf("| %s | %.2f/s | %.2f/s | %.1f%% |\n",
				mainSection.Test, mainSection.Time, prSection.Time, percentChange))
		} else { // No Change
			sb.WriteString(fmt.Sprintf("| %s | %.2f/s | %.2f/s | 0.0%% |\n",
				mainSection.Test, mainSection.Time, prSection.Time))
		}
	}
	fmt.Println(sb.String())
}
