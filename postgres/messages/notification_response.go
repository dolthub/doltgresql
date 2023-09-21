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
	initializeDefaultMessage(NotificationResponse{})
}

// NotificationResponse represents a PostgreSQL message.
type NotificationResponse struct {
	ProcessID int32
	Channel   string
	Payload   string
}

var notificationResponseDefault = Message{
	Name: "NotificationResponse",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('A'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "ProcessID",
			Type: Int32,
			Data: int32(0),
		},
		{
			Name: "Channel",
			Type: String,
			Data: "",
		},
		{
			Name: "Payload",
			Type: String,
			Data: "",
		},
	},
}

var _ MessageType = NotificationResponse{}

// encode implements the interface MessageType.
func (m NotificationResponse) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("ProcessID").MustWrite(m.ProcessID)
	outputMessage.Field("Channel").MustWrite(m.Channel)
	outputMessage.Field("Payload").MustWrite(m.Payload)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m NotificationResponse) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return NotificationResponse{
		ProcessID: s.Field("ProcessID").MustGet().(int32),
		Channel:   s.Field("Channel").MustGet().(string),
		Payload:   s.Field("Payload").MustGet().(string),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m NotificationResponse) defaultMessage() *Message {
	return &notificationResponseDefault
}
