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

package initialization

import (
	"sync"

	"github.com/dolthub/doltgresql/server/cast"
	"github.com/dolthub/doltgresql/server/config"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/binary"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/functions/unary"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var once = &sync.Once{}

// Initialize initializes each package across the project. This function should be used instead of an init() function.
func Initialize() {
	once.Do(func() {
		config.Init()
		pgtypes.InitBaseIDs()
		binary.Init()
		unary.Init()
		functions.Init()
		cast.Init()
		framework.Initialize()
	})
}
