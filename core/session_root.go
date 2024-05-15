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

package core

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
)

// UseRootInSession is a helper for modifying the RootValue from within functions, procedures, etc. If a new RootValue
// is returned, then it is written to the session. If a RootValue is not returned, then the session is not updated. A
// new RootValue should only be returned for changes to elements on the RootValue.
func UseRootInSession(ctx *sql.Context, f func(ctx *sql.Context, root *RootValue) (*RootValue, error)) error {
	session := dsess.DSessFromSess(ctx.Session)
	// Does this handle the current schema as well?
	state, ok, err := session.LookupDbState(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("UseRootInSession cannot find the database")
	}
	newRoot, err := f(ctx, state.WorkingRoot().(*RootValue))
	if err != nil {
		return err
	}
	if newRoot != nil {
		if err = session.SetWorkingRoot(ctx, ctx.GetCurrentDatabase(), newRoot); err != nil {
			return err
		}
	}
	return nil
}
