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

package id

// This adds all of the built-in schemas to the cache.
func init() {
	globalCache.setBuiltIn(NewId(Section_Namespace, "pg_catalog"), 11)
	globalCache.setBuiltIn(NewId(Section_Namespace, "pg_toast"), 99)
	globalCache.setBuiltIn(NewId(Section_Namespace, "public"), 2200)
	globalCache.setBuiltIn(NewId(Section_Namespace, "information_schema"), 13183)
}
