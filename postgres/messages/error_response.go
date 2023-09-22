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
	initializeDefaultMessage(ErrorResponse{})
}

// ErrorResponse represents a PostgreSQL message.
type ErrorResponse struct {
	Fields []ErrorResponseField
}

// ErrorResponseField are the fields to an ErrorResponse message.
type ErrorResponseField struct {
	Code  int32
	Value string
}

var errorResponseDefault = MessageFormat{
	Name: "ErrorResponse",
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

var _ Message = ErrorResponse{}

// encode implements the interface Message.
func (m ErrorResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	for i, field := range m.Fields {
		outputMessage.Field("Fields").Child("Code", i).MustWrite(field.Code)
		outputMessage.Field("Fields").Child("Value", i).MustWrite(field.Value)
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m ErrorResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("Fields").MustGet().(int32))
	fields := make([]ErrorResponseField, count)
	for i := 0; i < count; i++ {
		fields[i] = ErrorResponseField{
			Code:  s.Field("Fields").Child("Code", i).MustGet().(int32),
			Value: s.Field("Fields").Child("Value", i).MustGet().(string),
		}
	}
	return ErrorResponse{
		Fields: fields,
	}, nil
}

// defaultMessage implements the interface Message.
func (m ErrorResponse) defaultMessage() *MessageFormat {
	return &errorResponseDefault
}
