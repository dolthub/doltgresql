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
	"strings"
)

// MessageFormat is the format of a message as defined by PostgreSQL. Contains the description and values.
// https://www.postgresql.org/docs/15/protocol-message-formats.html
type MessageFormat struct {
	Name      string
	Fields    FieldGroup
	info      *messageInfo
	isDefault bool
}

// Message is a type that represents a PostgreSQL message.
type Message interface {
	// encode returns a new MessageFormat containing any modified data contained within the object. This should NOT be
	// the default message.
	encode() (MessageFormat, error)
	// decode returns a new Message that represents the given MessageFormat. You should never return the default
	// message, even if the message never varies from the default. Always make a copy, and then modify that copy.
	decode(s MessageFormat) (Message, error)
	// defaultMessage returns the default, unmodified message for this type.
	defaultMessage() *MessageFormat
}

// messageFieldInfo contains information on a specific field within a messageInfo.
type messageFieldInfo struct {
	RelativeIndex int
	Parent        string
	UsesByteCount bool // Only used by ByteN fields
}

// messageInfo contains all of the information that a message should keep track of. Used internally by messages.
type messageInfo struct {
	fieldInfo      map[string]messageFieldInfo
	appendNullByte bool
	defaultMessage *MessageFormat
}

// Copy returns a copy of the MessageFormat, which is free to modify.
func (m MessageFormat) Copy() MessageFormat {
	newFields := make(FieldGroup, len(m.Fields))
	for i, field := range m.Fields {
		newFields[i] = field.Copy()
	}
	return MessageFormat{m.Name, newFields, m.info, false}
}

// String returns a printable version of the MessageFormat.
func (m MessageFormat) String() string {
	buffer := strings.Builder{}
	buffer.WriteString(fmt.Sprintf("%s: {\n", m.Name))
	buffer.WriteString("\n") //TODO: print this
	buffer.WriteString("}")
	return buffer.String()
}

// MatchesStructure returns an error if the given MessageFormat has a different structure than the calling MessageFormat.
func (m MessageFormat) MatchesStructure(otherMessage MessageFormat) error {
	//TODO: check this
	return nil
}

// Field returns a MessageWriter for the calling MessageFormat, which makes it easier (and safer) to update the field whose
// name was given.
func (m MessageFormat) Field(name string) MessageWriter {
	return MessageWriter{
		message:    m,
		fieldQueue: []messageWriterChildPosition{{name, 0}},
	}
}
