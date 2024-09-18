package server

import (
	"context"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
)

type Handler interface {
	// ComBind is called when a connection receives a request to bind a prepared statement to a set of values.
	ComBind(ctx context.Context, c *mysql.Conn, query string, parsedQuery mysql.ParsedQuery, bindVars map[string]sqlparser.Expr) (mysql.BoundQuery, []pgproto3.FieldDescription, error)
	// ComExecuteBound is called when a connection receives a request to execute a prepared statement that has already bound to a set of values.
	ComExecuteBound(ctx context.Context, conn *mysql.Conn, query string, boundQuery mysql.BoundQuery, callback func(*Result) error) error
	// ComPrepareParsed is called when a connection receives a prepared statement query that has already been parsed.
	ComPrepareParsed(ctx context.Context, c *mysql.Conn, query string, parsed sqlparser.Statement) (mysql.ParsedQuery, []pgproto3.FieldDescription, error)
	// ComQuery is called when a connection receives a query. Note the contents of the query slice may change
	// after the first call to callback. So the DoltgresHandler should not hang on to the byte slice.
	ComQuery(ctx context.Context, c *mysql.Conn, query string, parsed sqlparser.Statement, callback func(*Result) error) error
	// ComResetConnection resets the connection's session, clearing out any cached prepared statements, locks, user and
	// session variables. The currently selected database is preserved.
	ComResetConnection(c *mysql.Conn) error
	// ConnectionClosed reports that a connection has been closed.
	ConnectionClosed(c *mysql.Conn)
	// NewConnection reports that a new connection has been established.
	NewConnection(c *mysql.Conn)
	// NewContext creates a new sql.Context instance for the connection |c|. The
	// optional |query| can be specified to populate the sql.Context's query field.
	NewContext(ctx context.Context, c *mysql.Conn, query string) (*sql.Context, error)
}
