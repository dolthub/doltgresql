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
	initializeDefaultMessage(PortalSuspended{})
}

// PortalSuspended represents a PostgreSQL message.
type PortalSuspended struct {
	Integer int32
	String  string
}

var portalSuspendedDefault = Message{
	Name: "PortalSuspended",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('s'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(4),
		},
	},
}

var _ MessageType = PortalSuspended{}

// encode implements the interface MessageType.
func (m PortalSuspended) encode() (Message, error) {
	return m.defaultMessage().Copy(), nil
}

// decode implements the interface MessageType.
func (m PortalSuspended) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return PortalSuspended{}, nil
}

// defaultMessage implements the interface MessageType.
func (m PortalSuspended) defaultMessage() *Message {
	return &portalSuspendedDefault
}
