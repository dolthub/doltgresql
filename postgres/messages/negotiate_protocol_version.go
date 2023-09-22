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
	connection.InitializeDefaultMessage(NegotiateProtocolVersion{})
}

// NegotiateProtocolVersion represents a PostgreSQL message.
type NegotiateProtocolVersion struct {
	NewestMinorProtocol int32
	UnrecognizedOptions []string
}

var negotiateProtocolVersionDefault = connection.MessageFormat{
	Name: "NegotiateProtocolVersion",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('v'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "NewestMinorProtocol",
			Type: connection.Int32,
			Data: int32(0),
		},
		{
			Name: "UnrecognizedOptions",
			Type: connection.Int32,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "UnrecognizedOption",
						Type: connection.String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ connection.Message = NegotiateProtocolVersion{}

// Encode implements the interface connection.Message.
func (m NegotiateProtocolVersion) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("NewestMinorProtocol").MustWrite(m.NewestMinorProtocol)
	for i, option := range m.UnrecognizedOptions {
		outputMessage.Field("UnrecognizedOptions").Child("UnrecognizedOption", i).MustWrite(option)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m NegotiateProtocolVersion) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
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

// DefaultMessage implements the interface connection.Message.
func (m NegotiateProtocolVersion) DefaultMessage() *connection.MessageFormat {
	return &negotiateProtocolVersionDefault
}
