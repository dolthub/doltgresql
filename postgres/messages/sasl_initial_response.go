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
	connection.InitializeDefaultMessage(SASLInitialResponse{})
}

// SASLInitialResponse represents a PostgreSQL message.
type SASLInitialResponse struct {
	Name     string
	Response []byte
}

var sASLInitialResponseDefault = connection.MessageFormat{
	Name: "SASLInitialResponse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('p'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Name",
			Type: connection.String,
			Data: "",
		},
		{
			Name:  "ResponseLength",
			Type:  connection.Int32,
			Flags: connection.ByteCount,
			Data:  int32(-1),
		},
		{
			Name: "ResponseData",
			Type: connection.String,
			Data: "",
		},
	},
}

var _ connection.Message = SASLInitialResponse{}

// Encode implements the interface connection.Message.
func (m SASLInitialResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("Name").MustWrite(m.Name)
	if len(m.Response) > 0 {
		outputMessage.Field("ResponseLength").MustWrite(len(m.Response))
		outputMessage.Field("ResponseData").MustWrite(m.Response)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m SASLInitialResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
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

// DefaultMessage implements the interface connection.Message.
func (m SASLInitialResponse) DefaultMessage() *connection.MessageFormat {
	return &sASLInitialResponseDefault
}
