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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeBeginTransaction handles *tree.BeginTransaction nodes.
func nodeBeginTransaction(ctx *Context, node *tree.BeginTransaction) (*vitess.Begin, error) {
	if node == nil {
		return nil, nil
	}
	if node.Modes.Isolation != tree.UnspecifiedIsolation {
		return nil, fmt.Errorf("isolation levels are not yet supported")
	}
	if node.Modes.UserPriority != tree.UnspecifiedUserPriority {
		return nil, fmt.Errorf("user priority is not yet supported")
	}
	if node.Modes.AsOf.Expr != nil {
		return nil, fmt.Errorf("AS OF is not yet supported")
	}
	if node.Modes.Deferrable != tree.UnspecifiedDeferrableMode {
		return nil, fmt.Errorf("deferrability is not yet supported")
	}
	var characteristic string
	switch node.Modes.ReadWriteMode {
	case tree.UnspecifiedReadWriteMode:
		characteristic = vitess.TxReadWrite
	case tree.ReadOnly:
		characteristic = vitess.TxReadOnly
	case tree.ReadWrite:
		characteristic = vitess.TxReadWrite
	default:
		return nil, fmt.Errorf("unknown READ/WRITE setting")
	}
	return &vitess.Begin{
		TransactionCharacteristic: characteristic,
	}, nil
}
