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
	initializeDefaultMessage(NoticeResponse{})
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

var noticeResponseDefault = MessageFormat{
	Name: "NoticeResponse",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('N'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name:  "Fields",
			Type:  Repeated,
			Flags: RepeatedTerminator,
			Data:  int32(0),
			Children: []FieldGroup{
				{
					{
						Name: "Code",
						Type: Byte1,
						Data: int32(0),
					},
					{
						Name: "Value",
						Type: String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ Message = NoticeResponse{}

// encode implements the interface Message.
func (m NoticeResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	for i, field := range m.Fields {
		outputMessage.Field("Fields").Child("Code", i).MustWrite(field.Code)
		outputMessage.Field("Fields").Child("Value", i).MustWrite(field.Value)
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m NoticeResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
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

// defaultMessage implements the interface Message.
func (m NoticeResponse) defaultMessage() *MessageFormat {
	return &noticeResponseDefault
}
