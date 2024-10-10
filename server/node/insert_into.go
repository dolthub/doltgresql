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

package node

//import (
//	"github.com/dolthub/go-mysql-server/sql"
//	"github.com/dolthub/go-mysql-server/sql/plan"
//	"github.com/dolthub/go-mysql-server/sql/rowexec"
//)
//
//// InsertInto is a node that implements functionality specifically relevant to Doltgres' table creation needs.
//type InsertInto struct {
//	gmsInsertInto *plan.InsertInto
//	domainCols    []DomainCol
//}
//
//var _ sql.ExecSourceRel = (*InsertInto)(nil)
//
//// NewInsertInto returns a new *InsertInto.
//func NewInsertInto(createTable *plan.InsertInto) *InsertInto {
//	return &InsertInto{
//		gmsInsertInto: createTable,
//	}
//}
//
//// CheckPrivileges implements the interface sql.ExecSourceRel.
//func (c *InsertInto) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
//	return c.gmsInsertInto.CheckPrivileges(ctx, opChecker)
//}
//
//// Children implements the interface sql.ExecSourceRel.
//func (c *InsertInto) Children() []sql.Node {
//	return c.gmsInsertInto.Children()
//}
//
//// IsReadOnly implements the interface sql.ExecSourceRel.
//func (c *InsertInto) IsReadOnly() bool {
//	return false
//}
//
//// Resolved implements the interface sql.ExecSourceRel.
//func (c *InsertInto) Resolved() bool {
//	return c.gmsInsertInto != nil && c.gmsInsertInto.Resolved()
//}
//
//// RowIter implements the interface sql.ExecSourceRel.
//func (c *InsertInto) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
//	createTableIter, err := rowexec.DefaultBuilder.Build(ctx, c.gmsInsertInto, r)
//	if err != nil {
//		return nil, err
//	}
//
//	// TODO: replace the check VALUE with column
//	// TODO: append domain checks to c.gmsInsertInto.Checks() and set it with c.gmsInsertInto.WithChecks()
//
//	return createTableIter, err
//}
//
//// Schema implements the interface sql.ExecSourceRel.
//func (c *InsertInto) Schema() sql.Schema {
//	return c.gmsInsertInto.Schema()
//}
//
//// String implements the interface sql.ExecSourceRel.
//func (c *InsertInto) String() string {
//	return c.gmsInsertInto.String()
//}
//
//// WithChildren implements the interface sql.ExecSourceRel.
//func (c *InsertInto) WithChildren(children ...sql.Node) (sql.Node, error) {
//	gmsInsertInto, err := c.gmsInsertInto.WithChildren(children...)
//	if err != nil {
//		return nil, err
//	}
//	return &InsertInto{
//		gmsInsertInto: gmsInsertInto.(*plan.InsertInto),
//		sequences:     c.sequences,
//	}, nil
//}
//
//type DomainCol struct {
//	ColName *sql.Column // todo idk
//	Checks  []sql.CheckConstraints
//}
