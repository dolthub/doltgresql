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
	initializeDefaultMessage(Close{})
	addMessageHeader(Close{})
}

// Close represents a PostgreSQL message.
type Close struct {
	ClosingPreparedStatement bool   // ClosingPreparedStatement: If true, closing a prepared statement. If false, closing a portal.
	Target                   string // Target is the name of whatever we are closing.
}

var closeDefault = MessageFormat{
	Name: "Close",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('C'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ClosingTarget",
			Type: Byte1,
			Data: int32(0),
		},
		{
			Name: "TargetName",
			Type: String,
			Data: "",
		},
	},
}

var _ Message = Close{}

// encode implements the interface Message.
func (m Close) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	if m.ClosingPreparedStatement {
		outputMessage.Field("ClosingTarget").MustWrite('S')
	} else {
		outputMessage.Field("ClosingTarget").MustWrite('P')
	}
	outputMessage.Field("TargetName").MustWrite(m.Target)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m Close) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
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

// defaultMessage implements the interface Message.
func (m Close) defaultMessage() *MessageFormat {
	return &closeDefault
}
