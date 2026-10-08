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

package _go

import "testing"

// TestExtendedProtocolTransitions verifies extended-query object lifetimes, error recovery, and batch boundaries.
func TestExtendedProtocolTransitions(t *testing.T) {
	setup := []string{"CREATE TABLE mytable (i BIGINT PRIMARY KEY);"}
	RunMessageFlowTests(t, []MessageFlowTest{
		{
			Name:        "Bind without Parse discards later messages until Sync",
			SetUpScript: setup,
			Steps: []FlowStep{
				Bind{PreparedStatement: "missing", ExpectedErr: `prepared statement "missing" does not exist`, ExpectedErrCode: "26000"},
				Parse{Name: "skipped", Query: "INSERT INTO mytable VALUES (1)"},
				Sync{},
				SimpleQuery{Query: "SELECT 1", Expected: []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"1"}}}}},
				QueryOnOtherConnection{Query: "SELECT count(*) FROM mytable", Expected: [][]string{{"0"}}},
			},
		},
		{
			Name: "Describe unknown statement recovers at Sync",
			Steps: []FlowStep{
				Describe{Name: "missing", ExpectedErr: `prepared statement "missing" does not exist`, ExpectedErrCode: "26000"},
				Sync{},
				Parse{Name: "select", Query: "SELECT 2"},
				Bind{PreparedStatement: "select"},
				Execute{Tag: "SELECT 1", Rows: [][]string{{"2"}}},
				Sync{},
			},
		},
		{
			Name: "Execute unknown portal recovers at Sync",
			Steps: []FlowStep{
				Execute{Portal: "missing", ExpectedErr: `portal "missing" does not exist`, ExpectedErrCode: "34000"},
				Sync{},
				SimpleQuery{Query: "SELECT 3", Expected: []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"3"}}}}},
			},
		},
		{
			Name: "named statement supports Describe before Bind and reuse after Sync",
			Steps: []FlowStep{
				Parse{Name: "saved", Query: "SELECT 4"},
				Describe{ObjectType: 'S', Name: "saved"},
				Sync{},
				Bind{PreparedStatement: "saved"},
				Execute{Tag: "SELECT 1", Rows: [][]string{{"4"}}},
				Sync{},
			},
		},
		{
			Name: "named statements and portals support phase-grouped pipelining",
			Steps: []FlowStep{
				Parse{Name: "first_statement", Query: "SELECT 11"},
				Parse{Name: "second_statement", Query: "SELECT 22"},
				Describe{Name: "first_statement"},
				Describe{Name: "second_statement"},
				Bind{PreparedStatement: "first_statement", Portal: "first_portal"},
				Bind{PreparedStatement: "second_statement", Portal: "second_portal"},
				Execute{Portal: "first_portal", Tag: "SELECT 1", Rows: [][]string{{"11"}}},
				Execute{Portal: "second_portal", Tag: "SELECT 1", Rows: [][]string{{"22"}}},
				Sync{},
			},
		},
		{
			Name: "duplicate named statement discards later messages and preserves original",
			Steps: []FlowStep{
				Parse{Name: "saved", Query: "SELECT 5"},
				Sync{},
				Parse{Name: "saved", Query: "SELECT 6", ExpectedErr: `prepared statement "saved" already exists`, ExpectedErrCode: "42P05"},
				Bind{PreparedStatement: "saved"},
				Sync{},
				Bind{PreparedStatement: "saved"},
				Execute{Tag: "SELECT 1", Rows: [][]string{{"5"}}},
				Sync{},
			},
		},
		{
			Name: "duplicate named portal discards later messages until Sync",
			Steps: []FlowStep{
				Parse{Name: "saved", Query: "SELECT 6"},
				Bind{PreparedStatement: "saved", Portal: "portal"},
				Bind{PreparedStatement: "saved", Portal: "portal", ExpectedErr: `cursor "portal" already exists`, ExpectedErrCode: "42P03"},
				Execute{Portal: "portal"},
				Sync{},
				SimpleQuery{Query: "SELECT 6", Expected: []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"6"}}}}},
			},
		},
		{
			Name: "invalid Describe subtype discards later messages until Sync",
			Steps: []FlowStep{
				Describe{ObjectType: 'X', Name: "bad", ExpectedErr: "invalid DESCRIBE message subtype 88", ExpectedErrCode: "08P01"},
				Parse{Name: "skipped", Query: "SELECT 9"},
				Sync{},
				SimpleQuery{Query: "SELECT 9", Expected: []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"9"}}}}},
			},
		},
		{
			Name: "invalid Close subtype discards later messages until Sync",
			Steps: []FlowStep{
				Close{ObjectType: 'X', Name: "bad", ExpectedErr: "invalid CLOSE message subtype 88", ExpectedErrCode: "08P01"},
				Parse{Name: "skipped", Query: "SELECT 10"},
				Sync{},
				SimpleQuery{Query: "SELECT 10", Expected: []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"10"}}}}},
			},
		},
		{
			Name: "Flush delivers responses without ending the extended batch",
			Steps: []FlowStep{
				Parse{Name: "saved", Query: "SELECT 7"},
				Flush{},
				Describe{ObjectType: 'S', Name: "saved"},
				Bind{PreparedStatement: "saved"},
				Execute{Tag: "SELECT 1", Rows: [][]string{{"7"}}},
				Sync{},
			},
		},
		{
			Name: "Sync is accepted while ready",
			Steps: []FlowStep{
				Sync{},
				SimpleQuery{Query: "SELECT 8", Expected: []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"8"}}}}},
			},
		},
	})
}
