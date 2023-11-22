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
	"fmt"
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// implicitCommitStatements are a collection of statements that perform an implicit COMMIT before executing. Such
// statements cannot have their effects reversed by rolling back a transaction or rolling back to a savepoint.
// https://dev.mysql.com/doc/refman/8.0/en/implicit-commit.html
var implicitCommitStatements = []string{"ALTER EVENT", "ALTER FUNCTION", "ALTER PROCEDURE", "ALTER SERVER",
	"ALTER TABLE", "ALTER TABLESPACE", "ALTER VIEW", "CALL", "CREATE DATABASE", "CREATE EVENT", "CREATE FUNCTION",
	"CREATE INDEX", "CREATE PROCEDURE", "CREATE ROLE", "CREATE SERVER", "CREATE SPATIAL REFERENCE SYSTEM",
	"CREATE TABLE", "CREATE TABLESPACE", "CREATE TRIGGER", "CREATE VIEW", "DROP DATABASE", "DROP EVENT",
	"DROP FUNCTION", "DROP INDEX", "DROP PROCEDURE", "DROP ROLE", "DROP SERVER", "DROP SPATIAL REFERENCE SYSTEM",
	"DROP TABLE", "DROP TABLESPACE", "DROP TRIGGER", "DROP VIEW", "INSTALL PLUGIN", "RENAME TABLE", "TRUNCATE TABLE",
	"UNINSTALL PLUGIN", "ALTER USER", "CREATE USER", "DROP USER", "GRANT", "RENAME USER", "REVOKE", "SET PASSWORD",
	"BEGIN", "LOCK TABLES", "START TRANSACTION", "UNLOCK TABLES", "LOAD DATA", "START REPLICA", "STOP REPLICA",
	"RESET REPLICA", "CHANGE REPLICATION SOURCE TO", "CHANGE MASTER TO"}

// ImplicitlyCommits returns whether the given statement implicitly commits. Case-insensitive.
func ImplicitlyCommits(statement string) bool {
	statement = strings.ToUpper(strings.TrimSpace(statement))
	for _, commitPrefix := range implicitCommitStatements {
		if strings.HasPrefix(statement, commitPrefix) {
			return true
		}
	}
	return false
}

// HandleImplicitCommitStatement returns a statement that can reverse the given statement, such that it appears to have
// never executed. This only applies to statements that implicitly commit, as determined by ImplicitlyCommits.
func HandleImplicitCommitStatement(statement string) (reverseStatement string, handled bool) {
	s, err := parser.Parse(statement)
	if err != nil || len(s) != 1 {
		return "", false
	}
	switch node := s[0].AST.(type) {
	case *tree.CreateDatabase:
		return fmt.Sprintf("DROP DATABASE %s", string(node.Name)), true
	case *tree.CreateTable:
		return fmt.Sprintf("DROP TABLE %s", node.Table.String()), true
	case *tree.CreateView:
		return fmt.Sprintf("DROP VIEW %s", node.Name.String()), true
	default:
		return "", false
	}
}
