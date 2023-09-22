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
	initializeDefaultMessage(Execute{})
	addMessageHeader(Execute{})
}

// Execute represents a PostgreSQL message.
type Execute struct {
	Portal string
	RowMax int32
}

var executeDefault = MessageFormat{
	Name: "Execute",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('E'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Portal",
			Type: String,
			Data: "",
		},
		{
			Name: "RowMax",
			Type: Int32,
			Data: int32(0),
		},
	},
}

var _ Message = Execute{}

// encode implements the interface Message.
func (m Execute) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("Portal").MustWrite(m.Portal)
	outputMessage.Field("RowMax").MustWrite(m.RowMax)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m Execute) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return Execute{
		Portal: s.Field("Portal").MustGet().(string),
		RowMax: s.Field("RowMax").MustGet().(int32),
	}, nil
}

// defaultMessage implements the interface Message.
func (m Execute) defaultMessage() *MessageFormat {
	return &executeDefault
}
