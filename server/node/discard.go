package node

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
)

// DiscardStatement is just a marker type, since all functionality is handled by the connection handler,
// rather than the engine. It has to conform to the sql.ExecSourceRel interface to be used in the handler, but this
// functionality is all unused.
type DiscardStatement struct{}

var _ sqlparser.Injectable = DiscardStatement{}
var _ sql.ExecSourceRel = DiscardStatement{}

func (d DiscardStatement) Resolved() bool {
	return true
}

func (d DiscardStatement) String() string {
	return "DISCARD ALL"
}

func (d DiscardStatement) Schema() sql.Schema {
	return nil
}

func (d DiscardStatement) Children() []sql.Node {
	return nil
}

func (d DiscardStatement) WithChildren(children ...sql.Node) (sql.Node, error) {
	return d, nil
}

func (d DiscardStatement) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

func (d DiscardStatement) IsReadOnly() bool {
	return true
}

func (d DiscardStatement) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	panic("DISCARD ALL should be handled by the connection handler")
}

func (d DiscardStatement) WithResolvedChildren(children []any) (any, error) {
	return d, nil
}
