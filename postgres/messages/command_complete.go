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
	"strconv"
	"strings"
)

func init() {
	initializeDefaultMessage(CommandComplete{})
	addMessageHeader(CommandComplete{})
}

// CommandComplete tells the client that the command has completed.
type CommandComplete struct {
	Query string
	Rows  int32
}

var commandCompleteDefault = Message{
	Name: "CommandComplete",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('C'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "CommandTag",
			Type: String,
			Data: "",
		},
	},
}

var _ MessageType = CommandComplete{}

// IsIUD returns whether the query is either an INSERT, UPDATE, or DELETE query.
func (m CommandComplete) IsIUD() bool {
	query := strings.TrimSpace(strings.ToLower(m.Query))
	if strings.HasPrefix(query, "insert") ||
		strings.HasPrefix(query, "update") ||
		strings.HasPrefix(query, "delete") {
		return true
	} else {
		return false
	}
}

// encode implements the interface MessageType.
func (m CommandComplete) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	query := strings.TrimSpace(strings.ToLower(m.Query))
	if strings.HasPrefix(query, "select") {
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("SELECT %d", m.Rows))
	} else if strings.HasPrefix(query, "insert") {
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("INSERT 0 %d", m.Rows))
	} else if strings.HasPrefix(query, "update") {
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("UPDATE %d", m.Rows))
	} else if strings.HasPrefix(query, "delete") {
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("DELETE %d", m.Rows))
	} else if strings.HasPrefix(query, "create") {
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("SELECT %d", m.Rows))
	} else if strings.HasPrefix(query, "call") {
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("SELECT %d", m.Rows))
	} else {
		return Message{}, fmt.Errorf("unsupported query for now")
	}
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m CommandComplete) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	query := strings.TrimSpace(s.Field("CommandTag").MustGet().(string))
	tokens := strings.Split(query, " ")
	rows, err := strconv.Atoi(tokens[len(tokens)-1])
	if err != nil {
		return nil, err
	}
	return CommandComplete{
		Query: query,
		Rows:  int32(rows),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m CommandComplete) defaultMessage() *Message {
	return &commandCompleteDefault
}
