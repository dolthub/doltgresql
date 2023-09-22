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
	initializeDefaultMessage(GSSResponse{})
}

// GSSResponse represents a PostgreSQL message.
type GSSResponse struct {
	Data []byte
}

var gSSResponseDefault = MessageFormat{
	Name: "GSSResponse",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('p'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Data",
			Type: ByteN,
			Data: []byte{},
		},
	},
}

var _ Message = GSSResponse{}

// encode implements the interface Message.
func (m GSSResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("Data").MustWrite(m.Data)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m GSSResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return GSSResponse{
		Data: s.Field("Data").MustGet().([]byte),
	}, nil
}

// defaultMessage implements the interface Message.
func (m GSSResponse) defaultMessage() *MessageFormat {
	return &gSSResponseDefault
}
