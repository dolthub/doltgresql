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
	initializeDefaultMessage(SASLInitialResponse{})
}

// SASLInitialResponse represents a PostgreSQL message.
type SASLInitialResponse struct {
	Name     string
	Response []byte
}

var sASLInitialResponseDefault = MessageFormat{
	Name: "SASLInitialResponse",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('p'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Name",
			Type: String,
			Data: "",
		},
		{
			Name:  "ResponseLength",
			Type:  Int32,
			Flags: ByteCount,
			Data:  int32(-1),
		},
		{
			Name: "ResponseData",
			Type: String,
			Data: "",
		},
	},
}

var _ Message = SASLInitialResponse{}

// encode implements the interface Message.
func (m SASLInitialResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("Name").MustWrite(m.Name)
	if len(m.Response) > 0 {
		outputMessage.Field("ResponseLength").MustWrite(len(m.Response))
		outputMessage.Field("ResponseData").MustWrite(m.Response)
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m SASLInitialResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	var responseData []byte
	if s.Field("ResponseLength").MustGet().(int32) > 0 {
		responseData = s.Field("ResponseData").MustGet().([]byte)
	}
	return SASLInitialResponse{
		Name:     s.Field("Name").MustGet().(string),
		Response: responseData,
	}, nil
}

// defaultMessage implements the interface Message.
func (m SASLInitialResponse) defaultMessage() *MessageFormat {
	return &sASLInitialResponseDefault
}
