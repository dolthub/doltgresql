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
	initializeDefaultMessage(StartupMessage{})
}

// StartupMessage is returned by the client upon connecting to the server, providing details about the client.
type StartupMessage struct {
	ProtocolMajorVersion int
	ProtocolMinorVersion int
	Parameters           map[string]string
}

var startupMessageDefault = Message{
	Name: "StartupMessage",
	Fields: []*Field{
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{ // The docs specify a single Int32 field, but the upper and lower bits are different values, so this just splits them
			Name: "ProtocolMajorVersion",
			Type: Int16,
			Data: int32(0),
		},
		{
			Name: "ProtocolMinorVersion",
			Type: Int16,
			Data: int32(0),
		},
		{
			Name: "Parameters",
			Type: Repeated,
			Tags: RepeatedTerminator,
			Data: int32(0),
			Children: [][]*Field{
				{
					{
						Name: "ParameterName",
						Type: String,
						Data: "",
					},
					{
						Name: "ParameterValue",
						Type: String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ MessageType = StartupMessage{}

// encode implements the interface MessageType.
func (m StartupMessage) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("ProtocolMajorVersion").MustWrite(m.ProtocolMajorVersion)
	outputMessage.Field("ProtocolMinorVersion").MustWrite(m.ProtocolMinorVersion)
	index := 0
	for name, value := range m.Parameters {
		outputMessage.Field("Parameters").Child("ParameterName", index).MustWrite(name)
		outputMessage.Field("Parameters").Child("ParameterValue", index).MustWrite(value)
		index++
	}
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m StartupMessage) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	parameters := make(map[string]string)
	count := int(s.Field("Parameters").MustGet().(int32))
	for i := 0; i < count; i++ {
		parameters[s.Field("Parameters").Child("ParameterName", i).MustGet().(string)] =
			s.Field("Parameters").Child("ParameterValue", i).MustGet().(string)
	}
	return StartupMessage{
		ProtocolMajorVersion: int(s.Field("ProtocolMajorVersion").MustGet().(int32)),
		ProtocolMinorVersion: int(s.Field("ProtocolMinorVersion").MustGet().(int32)),
		Parameters:           parameters,
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m StartupMessage) defaultMessage() *Message {
	return &startupMessageDefault
}
