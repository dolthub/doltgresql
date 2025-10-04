// Copyright 2025 Dolthub, Inc.
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

package ast

import (
	"os"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgnodes "github.com/dolthub/doltgresql/server/node"
)

const ignoreUnsupportedEnvKey = "DOLTGRES_IGNORE_UNSUPPORTED"

// ignoreUnsupportedStatements is a flag that determines whether to ignore unsupported statements. This is useful
// when importing a dump from postgres using certain import tools that expect every statement to succeed, including
// ones that we can't yet fully support (or that we never will, but are safe to ignore).
var ignoreUnsupportedStatements bool

func init() {
	if _, ignoreUnsupported := os.LookupEnv(ignoreUnsupportedEnvKey); ignoreUnsupported {
		ignoreUnsupportedStatements = true
	}
}

// NewNoOp returns a new NoOp statement which does nothing and issues zero or more warnings when run.
// Used for statements that aren't directly supported but which we don't want to cause errors.
func NewNoOp(warnings ...string) vitess.InjectedStatement {
	return vitess.InjectedStatement{
		Statement: pgnodes.NoOp{
			Warnings: warnings,
		},
	}
}

// NotYetSupportedError returns an unsupported error with the given message, or a NoOp statement if the environment
// variable DOLTGRES_IGNORE_UNSUPPORTED is set.
func NotYetSupportedError(errorMsg string) (vitess.Statement, error) {
	if ignoreUnsupportedStatements {
		return NewNoOp(errorMsg), nil
	}

	return nil, errors.New(errorMsg)
}
