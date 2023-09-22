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
	initializeDefaultMessage(EmptyQueryResponse{})
	addMessageHeader(EmptyQueryResponse{})
}

// EmptyQueryResponse represents a PostgreSQL message.
type EmptyQueryResponse struct{}

var emptyQueryResponseDefault = MessageFormat{
	Name: "EmptyQueryResponse",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('I'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(4),
		},
	},
}

var _ Message = EmptyQueryResponse{}

// encode implements the interface Message.
func (m EmptyQueryResponse) encode() (MessageFormat, error) {
	return m.defaultMessage().Copy(), nil
}

// decode implements the interface Message.
func (m EmptyQueryResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return EmptyQueryResponse{}, nil
}

// defaultMessage implements the interface Message.
func (m EmptyQueryResponse) defaultMessage() *MessageFormat {
	return &emptyQueryResponseDefault
}
