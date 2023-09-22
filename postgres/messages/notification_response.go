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
	connection.InitializeDefaultMessage(NotificationResponse{})
}

// NotificationResponse represents a PostgreSQL message.
type NotificationResponse struct {
	ProcessID int32
	Channel   string
	Payload   string
}

var notificationResponseDefault = connection.MessageFormat{
	Name: "NotificationResponse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('A'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ProcessID",
			Type: connection.Int32,
			Data: int32(0),
		},
		{
			Name: "Channel",
			Type: connection.String,
			Data: "",
		},
		{
			Name: "Payload",
			Type: connection.String,
			Data: "",
		},
	},
}

var _ connection.Message = NotificationResponse{}

// Encode implements the interface connection.Message.
func (m NotificationResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("ProcessID").MustWrite(m.ProcessID)
	outputMessage.Field("Channel").MustWrite(m.Channel)
	outputMessage.Field("Payload").MustWrite(m.Payload)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m NotificationResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return NotificationResponse{
		ProcessID: s.Field("ProcessID").MustGet().(int32),
		Channel:   s.Field("Channel").MustGet().(string),
		Payload:   s.Field("Payload").MustGet().(string),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m NotificationResponse) DefaultMessage() *connection.MessageFormat {
	return &notificationResponseDefault
}
