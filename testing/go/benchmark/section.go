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
	"strconv"
	"strings"
)

// Section represents a benchmark section.
type Section struct {
	Test string
	Time float64 // This is in number of iterations per second
}

// SectionResults creates a section for each test.
func SectionResults(fileData string) map[string]Section {
	sections := make(map[string]Section)
	for {
		headerStartIdx := strings.Index(fileData, `----`)
		if headerStartIdx == -1 {
			break
		}
		headerEndIdx := strings.Index(fileData[headerStartIdx+4:], `----`) + headerStartIdx + 4
		if headerEndIdx == headerStartIdx+4 {
			break
		}
		headerFull := fileData[headerStartIdx : headerEndIdx+4]
		endingHeaderIdx := strings.LastIndex(fileData, headerFull)
		if endingHeaderIdx == -1 {
			break
		}
		section := Section{
			Test: headerFull[4 : len(headerFull)-4],
			Time: -1,
		}
		sectionText := strings.TrimSpace(fileData[len(headerFull):endingHeaderIdx])
		fileData = fileData[endingHeaderIdx+len(headerFull):]
		for _, line := range strings.Split(sectionText, "\n") {
			if strings.Contains(line, `queries:`) {
				parenIdx := strings.Index(line, `(`)
				perSecIdx := strings.Index(line, ` per sec.)`)
				if parenIdx != -1 && perSecIdx != -1 {
					timeString := line[parenIdx+1 : perSecIdx]
					parsedTime, err := strconv.ParseFloat(timeString, 64)
					if err == nil {
						section.Time = parsedTime
					}
				}
				break
			}
		}
		if section.Time == -1 {
			continue
		}
		sections[section.Test] = section
	}
	return sections
}
