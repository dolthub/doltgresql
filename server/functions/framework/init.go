package framework

import (
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func Init() {
	pgtypes.IoOutput = IoOutput
	pgtypes.IoReceive = IoReceive
	pgtypes.IoSend = IoSend
	pgtypes.IoCompare = IoCompare
	pgtypes.SQL = SQL
}
