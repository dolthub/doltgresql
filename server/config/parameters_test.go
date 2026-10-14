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

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

func TestParameterInitValueUnsupportedUnit(t *testing.T) {
	for _, unit := range []string{"s", "kB", "MS"} {
		t.Run(unit, func(t *testing.T) {
			parameter := Parameter{Name: "unsupported_unit", Unit: unit}
			_, err := parameter.InitValue(nil, int64(1), false)
			require.EqualError(t, err, `unsupported unit "`+unit+`" for parameter "unsupported_unit"`)
			require.Equal(t, pgcode.FeatureNotSupported, pgerror.GetPGCode(err))
		})
	}
}
