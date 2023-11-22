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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(Bind{})
	connection.AddMessageHeader(Bind{})
}

// Bind represents a PostgreSQL message.
type Bind struct {
	DestinationPortal       string
	SourcePreparedStatement string
	ParameterFormatCodes    []int32
	ParameterValues         []BindParameterValue
	ResultFormatCodes       []int32
}

var _ sql.DebugStringer = Bind{}

// BindParameterValue are parameter values for the Bind message.
type BindParameterValue struct {
	Data   []byte
	IsNull bool
}

var bindDefault = connection.MessageFormat{
	Name: "Bind",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('B'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "DestinationPortal",
			Type: connection.String,
			Data: "",
		},
		{
			Name: "SourcePreparedStatement",
			Type: connection.String,
			Data: "",
		},
		{
			Name: "ParameterFormatCodes",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "ParameterFormatCode",
						Type: connection.Int16,
						Data: int32(0),
					},
				},
			},
		},
		{
			Name: "ParameterValues",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name:  "ParameterLength",
						Type:  connection.Int32,
						Flags: connection.ByteCount,
						Data:  int32(0),
					},
					{
						Name: "ParameterValue",
						Type: connection.ByteN,
						Data: []byte{},
					},
				},
			},
		},
		{
			Name: "ResultFormatCodes",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "ResultFormatCode",
						Type: connection.Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ connection.Message = Bind{}

// Encode implements the interface connection.Message.
func (m Bind) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("DestinationPortal").MustWrite(m.DestinationPortal)
	outputMessage.Field("SourcePreparedStatement").MustWrite(m.SourcePreparedStatement)
	for i, pFormatCode := range m.ParameterFormatCodes {
		outputMessage.Field("ParameterFormatCodes").Child("ParameterFormatCode", i).MustWrite(pFormatCode)
	}
	for i, paramValue := range m.ParameterValues {
		if paramValue.IsNull {
			outputMessage.Field("ParameterValues").Child("ParameterLength", i).MustWrite(-1)
		} else {
			outputMessage.Field("ParameterValues").Child("ParameterLength", i).MustWrite(len(paramValue.Data))
			outputMessage.Field("ParameterValues").Child("ParameterValue", i).MustWrite(paramValue.Data)
		}
	}
	for i, rFormatCode := range m.ResultFormatCodes {
		outputMessage.Field("ResultFormatCodes").Child("ResultFormatCode", i).MustWrite(rFormatCode)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m Bind) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}

	// Get the parameter format codes
	parameterFormatCodesCount := int(s.Field("ParameterFormatCodes").MustGet().(int32))
	parameterFormatCodes := make([]int32, parameterFormatCodesCount)
	for i := 0; i < parameterFormatCodesCount; i++ {
		parameterFormatCodes[i] = s.Field("ParameterFormatCodes").Child("ParameterFormatCode", i).MustGet().(int32)
	}
	// Get the parameter values
	parameterValuesCount := int(s.Field("ParameterValues").MustGet().(int32))
	parameterValues := make([]BindParameterValue, parameterValuesCount)
	for i := 0; i < parameterValuesCount; i++ {
		paramLength := s.Field("ParameterValues").Child("ParameterLength", i).MustGet().(int32)
		if paramLength == -1 {
			parameterValues[i] = BindParameterValue{
				IsNull: true,
			}
		} else {
			parameterValues[i] = BindParameterValue{
				Data:   s.Field("ParameterValues").Child("ParameterValue", i).MustGet().([]byte),
				IsNull: false,
			}
		}
	}
	// Get the result format codes
	resultFormatCodesCount := int(s.Field("ResultFormatCodes").MustGet().(int32))
	resultFormatCodes := make([]int32, resultFormatCodesCount)
	for i := 0; i < resultFormatCodesCount; i++ {
		resultFormatCodes[i] = s.Field("ResultFormatCodes").Child("ResultFormatCode", i).MustGet().(int32)
	}

	return Bind{
		DestinationPortal:       s.Field("DestinationPortal").MustGet().(string),
		SourcePreparedStatement: s.Field("SourcePreparedStatement").MustGet().(string),
		ParameterFormatCodes:    parameterFormatCodes,
		ParameterValues:         parameterValues,
		ResultFormatCodes:       resultFormatCodes,
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m Bind) DefaultMessage() *connection.MessageFormat {
	return &bindDefault
}

// DebugString returns a debug representation of the Bind message.
func (m Bind) DebugString() string {
	var builder strings.Builder

	builder.WriteString("Bind {\n")
	builder.WriteString(fmt.Sprintf("  DestinationPortal: %s\n", m.DestinationPortal))
	builder.WriteString(fmt.Sprintf("  SourcePreparedStatement: %s\n", m.SourcePreparedStatement))
	builder.WriteString("  ParameterFormatCodes: [")
	for _, code := range m.ParameterFormatCodes {
		builder.WriteString(fmt.Sprintf("%d, ", code))
	}
	builder.WriteString("]\n")

	builder.WriteString("  ParameterValues: [")
	for _, param := range m.ParameterValues {
		// Modify this part based on the structure of BindParameterValue.
		builder.WriteString(fmt.Sprintf("%#v, ", param))
	}
	builder.WriteString("]\n")

	builder.WriteString("  ResultFormatCodes: [")
	for _, code := range m.ResultFormatCodes {
		builder.WriteString(fmt.Sprintf("%d, ", code))
	}
	builder.WriteString("]\n")

	builder.WriteString("}")

	return builder.String()
}
