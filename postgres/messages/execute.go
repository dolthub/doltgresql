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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(Execute{})
	connection.AddMessageHeader(Execute{})
}

// Execute represents a PostgreSQL message.
type Execute struct {
	Portal string
	RowMax int32
}

var _ sql.DebugStringer = Execute{}

var executeDefault = connection.MessageFormat{
	Name: "Execute",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('E'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Portal",
			Type: connection.String,
			Data: "",
		},
		{
			Name: "RowMax",
			Type: connection.Int32,
			Data: int32(0),
		},
	},
}

var _ connection.Message = Execute{}

// Encode implements the interface connection.Message.
func (m Execute) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("Portal").MustWrite(m.Portal)
	outputMessage.Field("RowMax").MustWrite(m.RowMax)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m Execute) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return Execute{
		Portal: s.Field("Portal").MustGet().(string),
		RowMax: s.Field("RowMax").MustGet().(int32),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m Execute) DefaultMessage() *connection.MessageFormat {
	return &executeDefault
}

func (m Execute) DebugString() string {
	return fmt.Sprintf("Execute {\n  Portal: %s\n  RowMax: %d\n}", m.Portal, m.RowMax)
}
