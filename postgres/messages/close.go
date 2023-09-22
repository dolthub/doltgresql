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
	connection.InitializeDefaultMessage(Close{})
	connection.AddMessageHeader(Close{})
}

// Close represents a PostgreSQL message.
type Close struct {
	ClosingPreparedStatement bool   // ClosingPreparedStatement: If true, closing a prepared statement. If false, closing a portal.
	Target                   string // Target is the name of whatever we are closing.
}

var closeDefault = connection.MessageFormat{
	Name: "Close",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('C'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ClosingTarget",
			Type: connection.Byte1,
			Data: int32(0),
		},
		{
			Name: "TargetName",
			Type: connection.String,
			Data: "",
		},
	},
}

var _ connection.Message = Close{}

// Encode implements the interface connection.Message.
func (m Close) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	if m.ClosingPreparedStatement {
		outputMessage.Field("ClosingTarget").MustWrite('S')
	} else {
		outputMessage.Field("ClosingTarget").MustWrite('P')
	}
	outputMessage.Field("TargetName").MustWrite(m.Target)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m Close) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	closingTarget := s.Field("ClosingTarget").MustGet().(int32)
	var closingPreparedStatement bool
	if closingTarget == 'S' {
		closingPreparedStatement = true
	} else if closingTarget == 'P' {
		closingPreparedStatement = false
	} else {
		return nil, fmt.Errorf("Unknown closing target in Close message: %d", closingTarget)
	}
	return Close{
		ClosingPreparedStatement: closingPreparedStatement,
		Target:                   s.Field("TargetName").MustGet().(string),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m Close) DefaultMessage() *connection.MessageFormat {
	return &closeDefault
}
