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

package enginetest

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestServerLogging(t *testing.T) {
	for _, kind := range []string{"server", "query engine"} {
		t.Run(kind, func(t *testing.T) {
			previous := logrus.GetLevel()
			t.Cleanup(func() { logrus.SetLevel(previous) })
			logrus.SetLevel(logrus.DebugLevel)
			if kind == "server" {
				controller, _ := startServer(t, "localhost", "")
				t.Cleanup(func() {
					controller.Stop()
					require.NoError(t, controller.WaitForStop())
				})
			} else {
				engine := NewDoltgresQueryEngine(t, nil)
				t.Cleanup(func() { require.NoError(t, engine.Close()) })
			}
			require.Equal(t, logrus.WarnLevel, logrus.GetLevel())
			require.False(t, logrus.IsLevelEnabled(logrus.InfoLevel))
			require.True(t, logrus.IsLevelEnabled(logrus.WarnLevel))
			require.True(t, logrus.IsLevelEnabled(logrus.ErrorLevel))
		})
	}
}
