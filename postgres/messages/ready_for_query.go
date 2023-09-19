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

func init() {
	initializeDefaultMessage(ReadyForQuery{})
	addMessageHeader(ReadyForQuery{})
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

var readyForQueryDefault = Message{
	Name: "ReadyForQuery",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('Z'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(5),
		},
		{
			Name: "TransactionIndicator",
			Type: Byte1,
			Data: int32(0),
		},
	},
}

var _ MessageType = ReadyForQuery{}

// encode implements the interface MessageType.
func (m ReadyForQuery) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("TransactionIndicator").MustWrite(byte(m.Indicator))
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m ReadyForQuery) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return ReadyForQuery{
		Indicator: ReadyForQueryTransactionIndicator(s.Field("TransactionIndicator").MustGet().(int32)),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m ReadyForQuery) defaultMessage() *Message {
	return &readyForQueryDefault
}
