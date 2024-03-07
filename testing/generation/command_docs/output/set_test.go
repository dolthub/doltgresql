// Copyright 2023 Dolthub, Inc.
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

package output

import "testing"

func TestSet(t *testing.T) {
	tests := []QueryParses{
		Parses("SET configuration_parameter TO 1"),
		Parses("SET SESSION configuration_parameter TO 1"),
		Parses("SET LOCAL configuration_parameter TO 1"),
		Parses("SET configuration_parameter = 1"),
		Parses("SET SESSION configuration_parameter = 1"),
		Parses("SET LOCAL configuration_parameter = 1"),
		Parses("SET configuration_parameter TO ' 1 '"),
		Parses("SET SESSION configuration_parameter TO ' 1 '"),
		Parses("SET LOCAL configuration_parameter TO ' 1 '"),
		Parses("SET configuration_parameter = ' 1 '"),
		Parses("SET SESSION configuration_parameter = ' 1 '"),
		Parses("SET LOCAL configuration_parameter = ' 1 '"),
		Parses("SET configuration_parameter TO DEFAULT"),
		Parses("SET SESSION configuration_parameter TO DEFAULT"),
		Parses("SET LOCAL configuration_parameter TO DEFAULT"),
		Parses("SET configuration_parameter = DEFAULT"),
		Parses("SET SESSION configuration_parameter = DEFAULT"),
		Parses("SET LOCAL configuration_parameter = DEFAULT"),
		Converts("SET TIME ZONE 1"),
		Converts("SET SESSION TIME ZONE 1"),
		Parses("SET LOCAL TIME ZONE 1"),
		Converts("SET TIME ZONE ' 1 '"),
		Converts("SET SESSION TIME ZONE ' 1 '"),
		Parses("SET LOCAL TIME ZONE ' 1 '"),
		Converts("SET TIME ZONE LOCAL"),
		Converts("SET SESSION TIME ZONE LOCAL"),
		Parses("SET LOCAL TIME ZONE LOCAL"),
		Converts("SET TIME ZONE DEFAULT"),
		Converts("SET SESSION TIME ZONE DEFAULT"),
		Parses("SET LOCAL TIME ZONE DEFAULT"),
	}
	RunTests(t, tests)
}
