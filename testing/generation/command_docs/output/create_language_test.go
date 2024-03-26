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

func TestCreateLanguage(t *testing.T) {
	tests := []QueryParses{
		Parses("CREATE LANGUAGE name HANDLER call_handler"),
		Parses("CREATE OR REPLACE LANGUAGE name HANDLER call_handler"),
		Parses("CREATE TRUSTED LANGUAGE name HANDLER call_handler"),
		Parses("CREATE OR REPLACE TRUSTED LANGUAGE name HANDLER call_handler"),
		Parses("CREATE PROCEDURAL LANGUAGE name HANDLER call_handler"),
		Parses("CREATE OR REPLACE PROCEDURAL LANGUAGE name HANDLER call_handler"),
		Parses("CREATE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler"),
		Parses("CREATE OR REPLACE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler"),
		Parses("CREATE LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE OR REPLACE LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE TRUSTED LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE OR REPLACE TRUSTED LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE OR REPLACE PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE OR REPLACE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler"),
		Parses("CREATE LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE TRUSTED LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE TRUSTED LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE PROCEDURAL LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE PROCEDURAL LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler VALIDATOR valfunction"),
		Parses("CREATE LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE TRUSTED LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE TRUSTED LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE OR REPLACE TRUSTED PROCEDURAL LANGUAGE name HANDLER call_handler INLINE inline_handler VALIDATOR valfunction"),
		Parses("CREATE LANGUAGE name"),
		Parses("CREATE OR REPLACE LANGUAGE name"),
		Parses("CREATE TRUSTED LANGUAGE name"),
		Parses("CREATE OR REPLACE TRUSTED LANGUAGE name"),
		Parses("CREATE PROCEDURAL LANGUAGE name"),
		Parses("CREATE OR REPLACE PROCEDURAL LANGUAGE name"),
		Parses("CREATE TRUSTED PROCEDURAL LANGUAGE name"),
		Parses("CREATE OR REPLACE TRUSTED PROCEDURAL LANGUAGE name"),
	}
	RunTests(t, tests)
}
