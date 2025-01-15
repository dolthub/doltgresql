// Copyright 2025 Dolthub, Inc.
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

package id

import "testing"

// TestSectionValue operates as a line of defense to prevent accidental changes to pre-existing Section IDs.
// If this test fails, then a Section was changed that should not have been changed.
func TestSectionValue(t *testing.T) {
	ids := []struct {
		Section
		ID   uint8
		Name string
	}{
		{Section_Null, 0, "Null"},
		{Section_AccessMethod, 1, "AccessMethod"},
		{Section_Cast, 2, "Cast"},
		{Section_Check, 3, "Check"},
		{Section_Collation, 4, "Collation"},
		{Section_ColumnDefault, 5, "ColumnDefault"},
		{Section_Database, 6, "Database"},
		{Section_EnumLabel, 7, "EnumLabel"},
		{Section_EventTrigger, 8, "EventTrigger"},
		{Section_ExclusionConstraint, 9, "ExclusionConstraint"},
		{Section_Extension, 10, "Extension"},
		{Section_ForeignKey, 11, "ForeignKey"},
		{Section_ForeignDataWrapper, 12, "ForeignDataWrapper"},
		{Section_ForeignServer, 13, "ForeignServer"},
		{Section_ForeignTable, 14, "ForeignTable"},
		{Section_Function, 15, "Function"},
		{Section_FunctionLanguage, 16, "FunctionLanguage"},
		{Section_Index, 17, "Index"},
		{Section_Namespace, 18, "Namespace"},
		{Section_OID, 19, "OID"},
		{Section_Operator, 20, "Operator"},
		{Section_OperatorClass, 21, "OperatorClass"},
		{Section_OperatorFamily, 22, "OperatorFamily"},
		{Section_PrimaryKey, 23, "PrimaryKey"},
		{Section_Procedure, 24, "Procedure"},
		{Section_Publication, 25, "Publication"},
		{Section_RowLevelSecurity, 26, "RowLevelSecurity"},
		{Section_Sequence, 27, "Sequence"},
		{Section_Subscription, 28, "Subscription"},
		{Section_Table, 29, "Table"},
		{Section_TextSearchConfig, 30, "TextSearchConfig"},
		{Section_TextSearchDictionary, 31, "TextSearchDictionary"},
		{Section_TextSearchParser, 32, "TextSearchParser"},
		{Section_TextSearchTemplate, 33, "TextSearchTemplate"},
		{Section_Trigger, 34, "Trigger"},
		{Section_Type, 35, "Type"},
		{Section_UniqueKey, 36, "UniqueKey"},
		{Section_User, 37, "User"},
		{Section_View, 38, "View"},
	}
	allIds := make(map[uint8]string)
	for _, id := range ids {
		if uint8(id.Section) != id.ID {
			t.Logf("Section `%s` has been changed from its permanent value of `%d` to `%d`",
				id.Name, id.ID, uint8(id.Section))
			t.Fail()
		} else if existingName, ok := allIds[id.ID]; ok {
			t.Logf("Section `%s` has the same value as `%s`: `%d`",
				id.Name, existingName, id.ID)
			t.Fail()
		} else {
			allIds[id.ID] = id.Name
		}
	}
}
