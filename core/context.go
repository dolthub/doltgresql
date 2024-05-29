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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/sequences"
)

// contextValues contains a set of objects that will be passed alongside the context.
type contextValues struct {
	collection *sequences.Collection
}

// cvMap is a temporary map that holds context values, simply to get around the race condition of updating the context.
// TODO: remove this and add an actual construct to the context
var cvMap = map[*sql.Context]*contextValues{}

// getContextValues accesses the contextValues in the given context. If the context does not have a contextValues, then
// it creates one and adds it to the context.
func getContextValues(ctx *sql.Context) (*contextValues, error) {
	if cv, ok := cvMap[ctx]; ok {
		return cv, nil
	}
	cv := &contextValues{}
	cvMap[ctx] = cv
	return cv, nil
}

// getRootFromContext returns the working session's root from the context, along with the session.
func getRootFromContext(ctx *sql.Context) (*dsess.DoltSession, *RootValue, error) {
	session := dsess.DSessFromSess(ctx.Session)
	// Does this handle the current schema as well?
	state, ok, err := session.LookupDbState(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, fmt.Errorf("cannot find the database while fetching root from context")
	}
	return session, state.WorkingRoot().(*RootValue), nil
}

// GetTableFromContext returns the table from the context. Returns nil if no table was found.
func GetTableFromContext(ctx *sql.Context, tableName doltdb.TableName) (*doltdb.Table, error) {
	_, root, err := getRootFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var table *doltdb.Table
	if tableName.Schema == "" {
		_, table, _, err = resolve.Table(ctx, root, tableName.Name)
		if err != nil {
			return nil, err
		}
	} else {
		table, _, err = root.GetTable(ctx, tableName)
	}

	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetCollectionFromContext returns the given sequence collection from the context. Will always return a collection if
// no error is returned.
func GetCollectionFromContext(ctx *sql.Context) (*sequences.Collection, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.collection == nil {
		_, root, err := getRootFromContext(ctx)
		if err != nil {
			return nil, err
		}
		cv.collection, err = root.GetSequences(ctx)
		if err != nil {
			return nil, err
		}
	}
	return cv.collection, nil
}

// CloseContextRootFinalizer finalizes any changes persisted within the context by writing them to the working root.
// This should ONLY be called by the ContextRootFinalizer node.
func CloseContextRootFinalizer(ctx *sql.Context) error {
	cv, ok := cvMap[ctx]
	if !ok {
		return nil
	}
	if cv.collection == nil {
		return nil
	}
	session, root, err := getRootFromContext(ctx)
	if err != nil {
		return err
	}
	newRoot, err := root.PutSequences(ctx, cv.collection)
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
