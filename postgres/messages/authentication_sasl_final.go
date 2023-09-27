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
	connection.InitializeDefaultMessage(AuthenticationSASLFinal{})
}

// AuthenticationSASLFinal represents a PostgreSQL message.
type AuthenticationSASLFinal struct {
	AdditionalData []byte
}

var authenticationSASLFinalDefault = connection.MessageFormat{
	Name: "AuthenticationSASLFinal",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('R'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name:  "Status",
			Type:  connection.Int32,
			Flags: connection.StaticData,
			Data:  int32(12),
		},
		{
			Name: "AdditionalData",
			Type: connection.ByteN,
			Data: []byte{},
		},
	},
}

var _ connection.Message = AuthenticationSASLFinal{}

// Encode implements the interface connection.Message.
func (m AuthenticationSASLFinal) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("AdditionalData").MustWrite(m.AdditionalData)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m AuthenticationSASLFinal) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationSASLFinal{
		AdditionalData: s.Field("AdditionalData").MustGet().([]byte),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m AuthenticationSASLFinal) DefaultMessage() *connection.MessageFormat {
	return &authenticationSASLFinalDefault
}
