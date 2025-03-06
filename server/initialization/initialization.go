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

	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/servercfg"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	pgsql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/server/analyzer"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/cast"
	"github.com/dolthub/doltgresql/server/config"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/binary"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/functions/unary"
	"github.com/dolthub/doltgresql/server/plpgsql"
	"github.com/dolthub/doltgresql/server/tables"
	"github.com/dolthub/doltgresql/server/tables/dprocedures"
	"github.com/dolthub/doltgresql/server/tables/dtables"
	"github.com/dolthub/doltgresql/server/tables/information_schema"
	"github.com/dolthub/doltgresql/server/tables/pgcatalog"
	doltgresservercfg "github.com/dolthub/doltgresql/servercfg"
)

var once = &sync.Once{}

// Initialize initializes each package across the project. This function should be used instead of an init() function.
func Initialize(dEnv *env.DoltEnv) {
	once.Do(func() {
		core.Init()
		auth.Init(dEnv)
		pgexprs.Init()
		analyzer.Init()
		config.Init()
		binary.Init()
		unary.Init()
		functions.Init()
		cast.Init()
		framework.Initialize()
		sql.GlobalParser = pgsql.NewPostgresParser()
		servercfg.DefaultUnixSocketFilePath = doltgresservercfg.DefaultPostgresUnixSocketFilePath
		tables.Init()
		pgcatalog.Init()
		information_schema.Init()
		dtables.Init()
		dprocedures.Init()
		plpgsql.Init()
	})
}
