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

package types

import "github.com/lib/pq/oid"

// BpCharArray is the array variant of BpChar.
var BpCharArray = createArrayType(BpChar, SerializationID_CharArray, oid.T__bpchar)

// CharArray is the array variant of BpChar. This is an alias of BpCharArray, since the documentation references "char"
// more so than "bpchar" in PostgreSQL 15. They're the same type with different characteristics depending on the length.
var CharArray = createArrayType(InternalChar, SerializationID_CharArray, oid.T__char)
