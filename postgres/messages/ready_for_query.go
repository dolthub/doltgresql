package messages

// ReadyForQueryTransactionIndicator indicates the state of the transaction related to the query.
type ReadyForQueryTransactionIndicator byte

const (
	ReadyForQueryTransactionIndicator_Idle ReadyForQueryTransactionIndicator = iota
	ReadyForQueryTransactionIndicator_TransactionBlock
	ReadyForQueryTransactionIndicator_FailedTransactionBlock
)

// ReadyForQuery tells the client that the server is ready for a new query cycle.
type ReadyForQuery struct {
	Indicator ReadyForQueryTransactionIndicator
}

// Bytes returns ReadyForQuery as a byte slice, ready to be returned to the client.
func (rfq ReadyForQuery) Bytes() []byte {
	var indicator byte
	switch rfq.Indicator {
	case ReadyForQueryTransactionIndicator_Idle:
		indicator = 'I'
	case ReadyForQueryTransactionIndicator_TransactionBlock:
		indicator = 'T'
	case ReadyForQueryTransactionIndicator_FailedTransactionBlock:
		indicator = 'E'
	}
	return []byte{'Z', 0, 0, 0, 5, indicator}
}
