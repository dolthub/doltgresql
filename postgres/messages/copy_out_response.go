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

import "fmt"

func init() {
	initializeDefaultMessage(CopyOutResponse{})
}

// CopyOutResponse represents a PostgreSQL message.
type CopyOutResponse struct {
	IsTextual   bool // IsTextual states whether the copy is textual or binary.
	FormatCodes []int32
}

var copyOutResponseDefault = MessageFormat{
	Name: "CopyOutResponse",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('H'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ResponseType",
			Type: Int8,
			Data: int32(0),
		},
		{
			Name: "Columns",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name: "FormatCode",
						Type: Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ Message = CopyOutResponse{}

// encode implements the interface Message.
func (m CopyOutResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	if m.IsTextual {
		outputMessage.Field("ResponseType").MustWrite(0)
	} else {
		outputMessage.Field("ResponseType").MustWrite(1)
	}
	for i, formatCode := range m.FormatCodes {
		outputMessage.Field("Columns").Child("FormatCode", i).MustWrite(formatCode)
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m CopyOutResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	var isTextual bool
	responseType := s.Field("ResponseType").MustGet().(int32)
	if responseType == 0 {
		isTextual = true
	} else if responseType == 1 {
		isTextual = false
	} else {
		return nil, fmt.Errorf("Unknown response type in the CopyOutResponse message: %d", responseType)
	}
	count := int(s.Field("Columns").MustGet().(int32))
	formatCodes := make([]int32, count)
	for i := 0; i < count; i++ {
		formatCodes[i] = s.Field("Columns").Child("FormatCode", i).MustGet().(int32)
	}
	return CopyOutResponse{
		IsTextual:   isTextual,
		FormatCodes: formatCodes,
	}, nil
}

// defaultMessage implements the interface Message.
func (m CopyOutResponse) defaultMessage() *MessageFormat {
	return &copyOutResponseDefault
}
