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
	connection.InitializeDefaultMessage(AuthenticationSASL{})
}

// AuthenticationSASL represents a PostgreSQL message.
type AuthenticationSASL struct {
	Mechanisms []string
}

var authenticationSASLDefault = connection.MessageFormat{
	Name: "AuthenticationSASL",
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
			Data:  int32(10),
		},
		{
			Name:  "Mechanisms",
			Type:  connection.Repeated,
			Flags: connection.RepeatedTerminator,
			Data:  int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "Mechanism",
						Type: connection.String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ connection.Message = AuthenticationSASL{}

// Encode implements the interface connection.Message.
func (m AuthenticationSASL) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	for i, mechanism := range m.Mechanisms {
		outputMessage.Field("Mechanisms").Child("Mechanism", i).MustWrite(mechanism)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m AuthenticationSASL) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("Mechanisms").MustGet().(int32))
	mechanisms := make([]string, count)
	for i := 0; i < count; i++ {
		mechanisms[i] = s.Field("Mechanisms").Child("Mechanism", i).MustGet().(string)
	}
	return AuthenticationSASL{
		Mechanisms: mechanisms,
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m AuthenticationSASL) DefaultMessage() *connection.MessageFormat {
	return &authenticationSASLDefault
}
