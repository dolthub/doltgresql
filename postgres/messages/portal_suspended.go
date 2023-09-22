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

var portalSuspendedDefault = MessageFormat{
	Name: "PortalSuspended",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('s'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(4),
		},
	},
}

var _ Message = PortalSuspended{}

// encode implements the interface Message.
func (m PortalSuspended) encode() (MessageFormat, error) {
	return m.defaultMessage().Copy(), nil
}

// decode implements the interface Message.
func (m PortalSuspended) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return PortalSuspended{}, nil
}

// defaultMessage implements the interface Message.
func (m PortalSuspended) defaultMessage() *MessageFormat {
	return &portalSuspendedDefault
}
