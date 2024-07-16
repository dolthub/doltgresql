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

package oid

import pgtypes "github.com/dolthub/doltgresql/server/types"

// Init handles the assignment of the Io functions for the "reg" types.
func Init() {
	pgtypes.Regclass_IoInput = regclass_IoInput
	pgtypes.Regclass_IoOutput = regclass_IoOutput
	pgtypes.Regproc_IoInput = regproc_IoInput
	pgtypes.Regproc_IoOutput = regproc_IoOutput
	pgtypes.Regtype_IoInput = regtype_IoInput
	pgtypes.Regtype_IoOutput = regtype_IoOutput
}
