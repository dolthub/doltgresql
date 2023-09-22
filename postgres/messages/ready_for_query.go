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

package messages

import "github.com/dolthub/doltgresql/postgres/connection"

func init() {
	connection.InitializeDefaultMessage(ReadyForQuery{})
}

// ReadyForQueryTransactionIndicator indicates the state of the transaction related to the query.
type ReadyForQueryTransactionIndicator byte

const (
	ReadyForQueryTransactionIndicator_Idle                   ReadyForQueryTransactionIndicator = 'I'
	ReadyForQueryTransactionIndicator_TransactionBlock       ReadyForQueryTransactionIndicator = 'T'
	ReadyForQueryTransactionIndicator_FailedTransactionBlock ReadyForQueryTransactionIndicator = 'E'
)

// ReadyForQuery tells the client that the server is ready for a new query cycle.
type ReadyForQuery struct {
	Indicator ReadyForQueryTransactionIndicator
}

var readyForQueryDefault = connection.MessageFormat{
	Name: "ReadyForQuery",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('Z'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(5),
		},
		{
			Name: "TransactionIndicator",
			Type: connection.Byte1,
			Data: int32(0),
		},
	},
}

var _ connection.Message = ReadyForQuery{}

// Encode implements the interface connection.Message.
func (m ReadyForQuery) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("TransactionIndicator").MustWrite(byte(m.Indicator))
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m ReadyForQuery) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return ReadyForQuery{
		Indicator: ReadyForQueryTransactionIndicator(s.Field("TransactionIndicator").MustGet().(int32)),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m ReadyForQuery) DefaultMessage() *connection.MessageFormat {
	return &readyForQueryDefault
}
