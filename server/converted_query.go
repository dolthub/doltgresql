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

package server

import (
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
)

// ConvertedQuery represents a query that has been converted from the Postgres representation to the Vitess
// representation. String may contain the string version of the converted query. AST will contain the tree
// version of the converted query, and is the recommended form to use. If AST is nil, then use the String version,
// otherwise always prefer to AST.
type ConvertedQuery struct {
	String       string
	AST          vitess.Statement
	StatementTag string
}

type PreparedStatementData struct {
	Query        ConvertedQuery
	ReturnFields []pgproto3.FieldDescription
	BindVarTypes []uint32
}

type PortalData struct {
	Query        ConvertedQuery
	IsEmptyQuery bool
	Fields       []pgproto3.FieldDescription
	BoundPlan    sql.Node
}
