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

package functions

import "github.com/dolthub/doltgresql/server/functions/framework"

// initLog10 registers the functions to the catalog.
func initLog10() {
	log10_float64 := log_float64
	log10_numeric := log_numeric
	log10_float64.Name = "log10"
	log10_numeric.Name = "log10"
	log10_float64.Strict = true
	log10_numeric.Strict = true
	framework.RegisterFunction(log10_float64)
	framework.RegisterFunction(log10_numeric)
}
