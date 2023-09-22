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
	connection.InitializeDefaultMessage(NoticeResponse{})
}

// NoticeResponse represents a PostgreSQL message.
type NoticeResponse struct {
	Fields []NoticeResponseField
}

// NoticeResponseField are the fields to an NoticeResponse message.
type NoticeResponseField struct {
	Code  int32
	Value string
}

var noticeResponseDefault = connection.MessageFormat{
	Name: "NoticeResponse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('N'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name:  "Fields",
			Type:  connection.Repeated,
			Flags: connection.RepeatedTerminator,
			Data:  int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "Code",
						Type: connection.Byte1,
						Data: int32(0),
					},
					{
						Name: "Value",
						Type: connection.String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ connection.Message = NoticeResponse{}

// Encode implements the interface connection.Message.
func (m NoticeResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	for i, field := range m.Fields {
		outputMessage.Field("Fields").Child("Code", i).MustWrite(field.Code)
		outputMessage.Field("Fields").Child("Value", i).MustWrite(field.Value)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m NoticeResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("Fields").MustGet().(int32))
	fields := make([]NoticeResponseField, count)
	for i := 0; i < count; i++ {
		fields[i] = NoticeResponseField{
			Code:  s.Field("Fields").Child("Code", i).MustGet().(int32),
			Value: s.Field("Fields").Child("Value", i).MustGet().(string),
		}
	}
	return NoticeResponse{
		Fields: fields,
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m NoticeResponse) DefaultMessage() *connection.MessageFormat {
	return &noticeResponseDefault
}
