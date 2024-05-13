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

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(CommandComplete{})
}

// CommandComplete tells the client that the command has completed.
type CommandComplete struct {
	Query string
	Rows  int32
	Tag   string
}

var commandCompleteDefault = connection.MessageFormat{
	Name: "CommandComplete",
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
			Name: "CommandTag",
			Type: connection.String,
			Data: "",
		},
	},
}

var _ connection.Message = CommandComplete{}

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

// ReturnsRow returns whether the query returns set or rows such as SELECT and FETCH statements.
func (m CommandComplete) ReturnsRow() bool {
	switch m.Tag {
	case "SELECT", "SHOW", "FETCH":
		return true
	default:
		return false
	}
}

// Encode implements the interface connection.Message.
func (m CommandComplete) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	// https://www.postgresql.org/docs/current/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-COMMANDCOMPLETE
	switch m.Tag {
	case "INSERT", "DELETE", "UPDATE", "MERGE", "SELECT", "CREATE TABLE AS", "MOVE", "FETCH", "COPY":
		tag := m.Tag
		if tag == "INSERT" {
			tag = "INSERT 0"
		}
		outputMessage.Field("CommandTag").MustWrite(fmt.Sprintf("%s %d", tag, m.Rows))
	default:
		outputMessage.Field("CommandTag").MustWrite(m.Tag)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m CommandComplete) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
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
		Tag:   m.Tag,
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m CommandComplete) DefaultMessage() *connection.MessageFormat {
	return &commandCompleteDefault
}
