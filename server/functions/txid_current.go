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

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTxidCurrent registers the functions to the catalog.
func initTxidCurrent() {
	framework.RegisterFunction(txid_current)
}

// txid_current represents the PostgreSQL function of the same name, which returns the ID of the current transaction.
// IDs ascend in the order transactions start, for as long as the server runs.
var txid_current = framework.Function0{
	Name:               "txid_current",
	Return:             pgtypes.Int64,
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context) (any, error) {
		tx, ok := ctx.GetTransaction().(*dsess.DoltTransaction)
		if !ok {
			return int64(0), nil
		}

		return int64(tx.ID()), nil
	},
}
