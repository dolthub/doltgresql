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

package initialization

import (
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/config"
)

func TestNewPostgresEngineVariables(t *testing.T) {
	a, err := NewPostgresEngineVariables(nil, nil)
	require.NoError(t, err)
	b, err := NewPostgresEngineVariables(nil, nil)
	require.NoError(t, err)
	require.NotSame(t, a, b)

	aName, _, ok := a.System.GetGlobal("application_name")
	require.True(t, ok)
	bName, _, ok := b.System.GetGlobal("application_name")
	require.True(t, ok)
	require.IsType(t, &config.Parameter{}, aName)
	require.NotSame(t, aName, bName)

	_, _, ok = a.System.GetGlobal(dsess.DoltStatsEnabled)
	require.True(t, ok)
}
