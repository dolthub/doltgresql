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

package framework

import (
	"github.com/dolthub/doltgresql/server/types"
)

// Init handles the assignment of the IO functions for the types package.
func Init() {
	types.IoOutput = IoOutput
	types.IoReceive = IoReceive
	types.IoSend = IoSend
	types.IoCompare = IoCompare
	types.SQL = SQL
	types.TypModOut = TypModOut
}
