package _go

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScriptErrorDetail(t *testing.T) {
	err := &pgconn.PgError{Code: "2BP01", Message: "cannot drop table t", Detail: "function f(t) depends on type t"}
	require.Contains(t, scriptErrorText(err), "cannot drop table t\nfunction f(t) depends on type t")
	require.Contains(t, scriptErrorText(err), err.Error())
	require.Equal(t, "ordinary error", scriptErrorText(errors.New("ordinary error")))
	require.Equal(t, "2BP01", scriptErrorCode(err))
	require.Equal(t, "2BP01", scriptErrorCode(fmt.Errorf("prepare: %w", err)))
	require.Empty(t, scriptErrorCode(errors.New("client error")))
	require.Empty(t, scriptErrorCode(nil))
}
