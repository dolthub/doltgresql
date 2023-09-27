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
	connection.InitializeDefaultMessage(AuthenticationGSSContinue{})
}

// AuthenticationGSSContinue represents a PostgreSQL message.
type AuthenticationGSSContinue struct {
	Data []byte
}

var authenticationGSSContinueDefault = connection.MessageFormat{
	Name: "AuthenticationGSSContinue",
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
			Data:  int32(8),
		},
		{
			Name: "AuthenticationData",
			Type: connection.ByteN,
			Data: []byte{},
		},
	},
}

var _ connection.Message = AuthenticationGSSContinue{}

// Encode implements the interface connection.Message.
func (m AuthenticationGSSContinue) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("AuthenticationData").MustWrite(m.Data)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m AuthenticationGSSContinue) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationGSSContinue{
		Data: s.Field("AuthenticationData").MustGet().([]byte),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m AuthenticationGSSContinue) DefaultMessage() *connection.MessageFormat {
	return &authenticationGSSContinueDefault
}
