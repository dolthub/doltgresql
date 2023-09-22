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
	initializeDefaultMessage(NegotiateProtocolVersion{})
}

// NegotiateProtocolVersion represents a PostgreSQL message.
type NegotiateProtocolVersion struct {
	NewestMinorProtocol int32
	UnrecognizedOptions []string
}

var negotiateProtocolVersionDefault = MessageFormat{
	Name: "NegotiateProtocolVersion",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('v'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "NewestMinorProtocol",
			Type: Int32,
			Data: int32(0),
		},
		{
			Name: "UnrecognizedOptions",
			Type: Int32,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name: "UnrecognizedOption",
						Type: String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ Message = NegotiateProtocolVersion{}

// encode implements the interface Message.
func (m NegotiateProtocolVersion) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("NewestMinorProtocol").MustWrite(m.NewestMinorProtocol)
	for i, option := range m.UnrecognizedOptions {
		outputMessage.Field("UnrecognizedOptions").Child("UnrecognizedOption", i).MustWrite(option)
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m NegotiateProtocolVersion) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("UnrecognizedOptions").MustGet().(int32))
	unrecognizedOptions := make([]string, count)
	for i := 0; i < count; i++ {
		unrecognizedOptions[i] = s.Field("UnrecognizedOptions").Child("UnrecognizedOption", i).MustGet().(string)
	}
	return NegotiateProtocolVersion{
		NewestMinorProtocol: s.Field("NewestMinorProtocol").MustGet().(int32),
		UnrecognizedOptions: unrecognizedOptions,
	}, nil
}

// defaultMessage implements the interface Message.
func (m NegotiateProtocolVersion) defaultMessage() *MessageFormat {
	return &negotiateProtocolVersionDefault
}
