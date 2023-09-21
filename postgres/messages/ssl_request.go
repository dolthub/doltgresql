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
	initializeDefaultMessage(SSLRequest{})
}

// SSLRequest represents a PostgreSQL message.
type SSLRequest struct{}

var sslRequestDefault = Message{
	Name: "SSLRequest",
	Fields: []*Field{
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(8),
		},
		{
			Name: "RequestCode",
			Type: Int32,
			Data: int32(80877103),
		},
	},
}

var _ MessageType = SSLRequest{}

// encode implements the interface MessageType.
func (m SSLRequest) encode() (Message, error) {
	return m.defaultMessage().Copy(), nil
}

// decode implements the interface MessageType.
func (m SSLRequest) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return SSLRequest{}, nil
}

// defaultMessage implements the interface MessageType.
func (m SSLRequest) defaultMessage() *Message {
	return &sslRequestDefault
}
