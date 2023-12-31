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

func TestSetTransaction(t *testing.T) {
	tests := []QueryParses{
		Parses("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ COMMITTED"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET TRANSACTION READ WRITE"),
		Parses("SET TRANSACTION READ ONLY"),
		Parses("SET TRANSACTION DEFERRABLE"),
		Parses("SET TRANSACTION NOT DEFERRABLE"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET TRANSACTION READ WRITE , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET TRANSACTION READ ONLY , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET TRANSACTION DEFERRABLE , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET TRANSACTION READ WRITE , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET TRANSACTION READ ONLY , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET TRANSACTION DEFERRABLE , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET TRANSACTION READ WRITE , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET TRANSACTION READ ONLY , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET TRANSACTION DEFERRABLE , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL READ UNCOMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL READ UNCOMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL READ UNCOMMITTED"),
		Unimplemented("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET TRANSACTION READ WRITE , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET TRANSACTION READ ONLY , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET TRANSACTION DEFERRABLE , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , READ WRITE"),
		Parses("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , READ WRITE"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , READ WRITE"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , READ WRITE"),
		Unimplemented("SET TRANSACTION READ WRITE , READ WRITE"),
		Unimplemented("SET TRANSACTION READ ONLY , READ WRITE"),
		Parses("SET TRANSACTION DEFERRABLE , READ WRITE"),
		Parses("SET TRANSACTION NOT DEFERRABLE , READ WRITE"),
		Parses("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , READ ONLY"),
		Parses("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , READ ONLY"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , READ ONLY"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , READ ONLY"),
		Unimplemented("SET TRANSACTION READ WRITE , READ ONLY"),
		Unimplemented("SET TRANSACTION READ ONLY , READ ONLY"),
		Parses("SET TRANSACTION DEFERRABLE , READ ONLY"),
		Parses("SET TRANSACTION NOT DEFERRABLE , READ ONLY"),
		Parses("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , DEFERRABLE"),
		Parses("SET TRANSACTION READ WRITE , DEFERRABLE"),
		Parses("SET TRANSACTION READ ONLY , DEFERRABLE"),
		Unimplemented("SET TRANSACTION DEFERRABLE , DEFERRABLE"),
		Unimplemented("SET TRANSACTION NOT DEFERRABLE , DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE , NOT DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ , NOT DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ COMMITTED , NOT DEFERRABLE"),
		Parses("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , NOT DEFERRABLE"),
		Parses("SET TRANSACTION READ WRITE , NOT DEFERRABLE"),
		Parses("SET TRANSACTION READ ONLY , NOT DEFERRABLE"),
		Unimplemented("SET TRANSACTION DEFERRABLE , NOT DEFERRABLE"),
		Unimplemented("SET TRANSACTION NOT DEFERRABLE , NOT DEFERRABLE"),
		Unimplemented("SET TRANSACTION SNAPSHOT 'snapshot_id'"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , ISOLATION LEVEL SERIALIZABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL SERIALIZABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , ISOLATION LEVEL REPEATABLE READ"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL REPEATABLE READ"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , ISOLATION LEVEL READ COMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL READ COMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , ISOLATION LEVEL READ UNCOMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , ISOLATION LEVEL READ UNCOMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , ISOLATION LEVEL READ UNCOMMITTED"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , ISOLATION LEVEL READ UNCOMMITTED"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , READ WRITE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , READ WRITE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , READ WRITE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , READ ONLY"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , READ ONLY"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , READ ONLY"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , DEFERRABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , DEFERRABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE , NOT DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ , NOT DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED , NOT DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ UNCOMMITTED , NOT DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE , NOT DEFERRABLE"),
		Parses("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY , NOT DEFERRABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION DEFERRABLE , NOT DEFERRABLE"),
		Unimplemented("SET SESSION CHARACTERISTICS AS TRANSACTION NOT DEFERRABLE , NOT DEFERRABLE"),
	}
	RunTests(t, tests)
}
