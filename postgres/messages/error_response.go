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
	"strings"

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(ErrorResponse{})
}

// ErrorResponseSeverity represents the severity of an ErrorResponse message.
type ErrorResponseSeverity string

const (
	ErrorResponseSeverity_Error   ErrorResponseSeverity = "ERROR"
	ErrorResponseSeverity_Fatal   ErrorResponseSeverity = "FATAL"
	ErrorResponseSeverity_Panic   ErrorResponseSeverity = "PANIC"
	ErrorResponseSeverity_Warning ErrorResponseSeverity = "WARNING"
	ErrorResponseSeverity_Notice  ErrorResponseSeverity = "NOTICE"
	ErrorResponseSeverity_Debug   ErrorResponseSeverity = "DEBUG"
	ErrorResponseSeverity_Info    ErrorResponseSeverity = "INFO"
	ErrorResponseSeverity_Log     ErrorResponseSeverity = "LOG"
)

// ErrorResponse represents a server-side error that should be returned to the client. The Optional fields do not need
// to be set, but may give additional context for the error.
type ErrorResponse struct {
	Severity     ErrorResponseSeverity
	SqlStateCode string
	Message      string
	Optional     ErrorResponseOptionalFields
}

// ErrorResponseOptionalFields are optional fields that will not be sent if their values are empty strings.
type ErrorResponseOptionalFields struct {
	Schema     string
	Table      string
	Column     string
	Constraint string
	Routine    string
}

var errorResponseDefault = connection.MessageFormat{
	Name: "ErrorResponse",
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
			Name:  "Fields",
			Type:  connection.Repeated,
			Flags: connection.RepeatedTerminator,
			Data:  int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "Code",
						Type: connection.Byte1,
						Data: int32(0),
					},
					{
						Name: "Value",
						Type: connection.String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ connection.Message = ErrorResponse{}

// Encode implements the interface connection.Message.
func (m ErrorResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	// Write the required fields first
	outputMessage.Field("Fields").Child("Code", 0).MustWrite('S')
	outputMessage.Field("Fields").Child("Value", 0).MustWrite(string(m.Severity))
	outputMessage.Field("Fields").Child("Code", 1).MustWrite('V')
	outputMessage.Field("Fields").Child("Value", 1).MustWrite(string(m.Severity))
	outputMessage.Field("Fields").Child("Code", 2).MustWrite('C')
	outputMessage.Field("Fields").Child("Value", 2).MustWrite(m.SqlStateCode)
	outputMessage.Field("Fields").Child("Code", 3).MustWrite('M')
	outputMessage.Field("Fields").Child("Value", 3).MustWrite(m.Message)

	// Write the optional fields after the required fields
	i := 4
	if len(m.Optional.Schema) > 0 {
		outputMessage.Field("Fields").Child("Code", i).MustWrite('s')
		outputMessage.Field("Fields").Child("Value", i).MustWrite(m.Optional.Schema)
		i++
	}
	if len(m.Optional.Table) > 0 {
		outputMessage.Field("Fields").Child("Code", i).MustWrite('t')
		outputMessage.Field("Fields").Child("Value", i).MustWrite(m.Optional.Table)
		i++
	}
	if len(m.Optional.Column) > 0 {
		outputMessage.Field("Fields").Child("Code", i).MustWrite('c')
		outputMessage.Field("Fields").Child("Value", i).MustWrite(m.Optional.Column)
		i++
	}
	if len(m.Optional.Constraint) > 0 {
		outputMessage.Field("Fields").Child("Code", i).MustWrite('n')
		outputMessage.Field("Fields").Child("Value", i).MustWrite(m.Optional.Constraint)
		i++
	}
	if len(m.Optional.Routine) > 0 {
		outputMessage.Field("Fields").Child("Code", i).MustWrite('R')
		outputMessage.Field("Fields").Child("Value", i).MustWrite(m.Optional.Routine)
		i++
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m ErrorResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	errorResponse := ErrorResponse{}
	count := int(s.Field("Fields").MustGet().(int32))
	for i := 0; i < count; i++ {
		value := s.Field("Fields").Child("Value", i).MustGet().(string)
		switch s.Field("Fields").Child("Code", i).MustGet().(int32) {
		case 'S':
			errorResponse.Severity = ErrorResponseSeverity(strings.ToUpper(strings.TrimSpace(value)))
		case 'V':
			errorResponse.Severity = ErrorResponseSeverity(strings.ToUpper(strings.TrimSpace(value)))
		case 'C':
			errorResponse.SqlStateCode = value
		case 'M':
			errorResponse.Message = value
		case 's':
			errorResponse.Optional.Schema = value
		case 't':
			errorResponse.Optional.Table = value
		case 'c':
			errorResponse.Optional.Column = value
		case 'n':
			errorResponse.Optional.Constraint = value
		case 'R':
			errorResponse.Optional.Routine = value
		}
	}
	return errorResponse, nil
}

// DefaultMessage implements the interface connection.Message.
func (m ErrorResponse) DefaultMessage() *connection.MessageFormat {
	return &errorResponseDefault
}
