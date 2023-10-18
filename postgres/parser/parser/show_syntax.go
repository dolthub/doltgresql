// Copyright 2023 Dolthub, Inc.
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

// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import (
	"context"
	"fmt"

	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// RunShowSyntax analyzes the syntax and reports its structure as data
// for the client. Even an error is reported as data.
//
// Since errors won't propagate to the client as an error, but as
// a result, the usual code path to capture and record errors will not
// be triggered. Instead, the caller can pass a reportErr closure to
// capture errors instead. May be nil.
func RunShowSyntax(
	ctx context.Context,
	stmt string,
	report func(ctx context.Context, field, msg string),
	reportErr func(ctx context.Context, err error),
) {
	stmts, err := Parse(stmt)
	if err != nil {
		if reportErr != nil {
			reportErr(ctx, err)
		}

		pqErr := pgerror.Flatten(err)
		report(ctx, "error", pqErr.Message)
		report(ctx, "code", pqErr.Code)
		if pqErr.Source != nil {
			if pqErr.Source.File != "" {
				report(ctx, "file", pqErr.Source.File)
			}
			if pqErr.Source.Line > 0 {
				report(ctx, "line", fmt.Sprintf("%d", pqErr.Source.Line))
			}
			if pqErr.Source.Function != "" {
				report(ctx, "function", pqErr.Source.Function)
			}
		}
		if pqErr.Detail != "" {
			report(ctx, "detail", pqErr.Detail)
		}
		if pqErr.Hint != "" {
			report(ctx, "hint", pqErr.Hint)
		}
	} else {
		for i := range stmts {
			report(ctx, "sql", tree.AsStringWithFlags(stmts[i].AST, tree.FmtParsable))
		}
	}
}
