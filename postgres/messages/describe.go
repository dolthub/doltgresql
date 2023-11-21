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
	"github.com/dolthub/go-mysql-server/sql"
)

func init() {
	connection.InitializeDefaultMessage(Describe{})
	connection.AddMessageHeader(Describe{})
}

// Describe represents a PostgreSQL message.
type Describe struct {
	IsPrepared bool // IsPrepared states whether we're describing a prepared statement or a portal.
	Target     string
}

var _ sql.DebugStringer = Describe{}

var describeDefault = connection.MessageFormat{
	Name: "Describe",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('D'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "DescribingTarget",
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

var _ connection.Message = Describe{}

// Encode implements the interface connection.Message.
func (m Describe) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	if m.IsPrepared {
		outputMessage.Field("DescribingTarget").MustWrite('S')
	} else {
		outputMessage.Field("DescribingTarget").MustWrite('P')
	}
	outputMessage.Field("TargetName").MustWrite(m.Target)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m Describe) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	describingTarget := s.Field("DescribingTarget").MustGet().(int32)
	var isPrepared bool
	if describingTarget == 'S' {
		isPrepared = true
	} else if describingTarget == 'P' {
		isPrepared = false
	} else {
		return nil, fmt.Errorf("Unknown describing target in Describe message: %d", describingTarget)
	}
	return Describe{
		IsPrepared: isPrepared,
		Target:     s.Field("TargetName").MustGet().(string),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m Describe) DefaultMessage() *connection.MessageFormat {
	return &describeDefault
}

func (m Describe) DebugString() string {
	return fmt.Sprintf("Describe { IsPrepared: %v, Target: %s }", m.IsPrepared, m.Target)
}
