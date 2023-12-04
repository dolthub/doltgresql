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

func TestCreateSubscription(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name WITH ( subscription_parameter )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name WITH ( subscription_parameter )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name WITH ( subscription_parameter = value )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name WITH ( subscription_parameter = value )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name WITH ( subscription_parameter , subscription_parameter )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name WITH ( subscription_parameter , subscription_parameter )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name WITH ( subscription_parameter = value , subscription_parameter )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name WITH ( subscription_parameter = value , subscription_parameter )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name WITH ( subscription_parameter , subscription_parameter = value )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name WITH ( subscription_parameter , subscription_parameter = value )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name WITH ( subscription_parameter = value , subscription_parameter = value )"),
		Unimplemented("CREATE SUBSCRIPTION subscription_name CONNECTION ' conninfo ' PUBLICATION publication_name , publication_name WITH ( subscription_parameter = value , subscription_parameter = value )"),
	}
	RunTests(t, tests)
}
