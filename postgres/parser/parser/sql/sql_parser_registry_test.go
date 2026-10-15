package sql_test

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/server/config"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

func TestVariableStatementsRequireAnEngineRegistry(t *testing.T) {
	parser := psql.NewPostgresParser()
	_, err := parser.ParseSimple("SET application_name TO 'client'")
	require.ErrorContains(t, err, "engine variable registry")

	vars := sqle.NewEngineVariables()
	config.InitWithRegistry(vars.System)
	ctx := sql.NewContextWithVariables(context.Background(), vars)
	statement, _, _, err := parser.Parse(ctx, "SET application_name TO 'client'", false)
	require.NoError(t, err)
	require.NotNil(t, statement)
}
