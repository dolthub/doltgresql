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

import (
	"fmt"

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(CopyInResponse{})
}

// CopyInResponse represents a PostgreSQL message.
type CopyInResponse struct {
	IsTextual   bool // IsTextual states whether the copy is textual or binary.
	FormatCodes []int32
}

var copyInResponseDefault = connection.MessageFormat{
	Name: "CopyInResponse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('G'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ResponseType",
			Type: connection.Int8,
			Data: int32(0),
		},
		{
			Name: "Columns",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "FormatCode",
						Type: connection.Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ connection.Message = CopyInResponse{}

// Encode implements the interface connection.Message.
func (m CopyInResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
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

// Decode implements the interface connection.Message.
func (m CopyInResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	var isTextual bool
	responseType := s.Field("ResponseType").MustGet().(int32)
	if responseType == 0 {
		isTextual = true
	} else if responseType == 1 {
		isTextual = false
	} else {
		return nil, fmt.Errorf("Unknown response type in the CopyInResponse message: %d", responseType)
	}
	count := int(s.Field("Columns").MustGet().(int32))
	formatCodes := make([]int32, count)
	for i := 0; i < count; i++ {
		formatCodes[i] = s.Field("Columns").Child("FormatCode", i).MustGet().(int32)
	}
	return CopyInResponse{
		IsTextual:   isTextual,
		FormatCodes: formatCodes,
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m CopyInResponse) DefaultMessage() *connection.MessageFormat {
	return &copyInResponseDefault
}
