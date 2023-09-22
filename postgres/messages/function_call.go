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
	connection.InitializeDefaultMessage(FunctionCall{})
	connection.AddMessageHeader(FunctionCall{})
}

// FunctionCall represents a PostgreSQL message.
type FunctionCall struct {
	ObjectID            int32
	ArgumentFormatCodes []int32
	Arguments           []FunctionCallArgument
	ResultFormatCode    int32
}

// FunctionCallArgument are arguments for the FunctionCall message.
type FunctionCallArgument struct {
	Data   []byte
	IsNull bool
}

var functionCallDefault = connection.MessageFormat{
	Name: "FunctionCall",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('F'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ObjectID",
			Type: connection.Int32,
			Data: int32(0),
		},
		{
			Name: "ArgumentFormatCodes",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "ArgumentFormatCode",
						Type: connection.Int16,
						Data: int32(0),
					},
				},
			},
		},
		{
			Name: "Arguments",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name:  "ArgumentLength",
						Type:  connection.Int32,
						Flags: connection.ByteCount,
						Data:  int32(0),
					},
					{
						Name: "ArgumentValue",
						Type: connection.ByteN,
						Data: []byte{},
					},
				},
			},
		},
		{
			Name: "ResultFormatCode",
			Type: connection.Int16,
			Data: int32(0),
		},
	},
}

var _ connection.Message = FunctionCall{}

// Encode implements the interface connection.Message.
func (m FunctionCall) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("ObjectID").MustWrite(m.ObjectID)
	for i, formatCode := range m.ArgumentFormatCodes {
		outputMessage.Field("ArgumentFormatCodes").Child("ArgumentFormatCode", i).MustWrite(formatCode)
	}
	for i, argument := range m.Arguments {
		if argument.IsNull {
			outputMessage.Field("Arguments").Child("ArgumentLength", i).MustWrite(-1)
		} else {
			outputMessage.Field("Arguments").Child("ArgumentLength", i).MustWrite(len(argument.Data))
			outputMessage.Field("Arguments").Child("ArgumentValue", i).MustWrite(argument.Data)
		}
	}
	outputMessage.Field("ResultFormatCode").MustWrite(m.ResultFormatCode)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m FunctionCall) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}

	// Get the argument format codes
	argumentFormatCodesCount := int(s.Field("ArgumentFormatCodes").MustGet().(int32))
	argumentFormatCodes := make([]int32, argumentFormatCodesCount)
	for i := 0; i < argumentFormatCodesCount; i++ {
		argumentFormatCodes[i] = s.Field("ArgumentFormatCodes").Child("ArgumentFormatCode", i).MustGet().(int32)
	}
	// Get the arguments
	argumentsCount := int(s.Field("Arguments").MustGet().(int32))
	arguments := make([]FunctionCallArgument, argumentsCount)
	for i := 0; i < argumentsCount; i++ {
		paramLength := s.Field("Arguments").Child("ArgumentLength", i).MustGet().(int32)
		if paramLength == -1 {
			arguments[i] = FunctionCallArgument{
				IsNull: true,
			}
		} else {
			arguments[i] = FunctionCallArgument{
				Data:   s.Field("Arguments").Child("ArgumentValue", i).MustGet().([]byte),
				IsNull: false,
			}
		}
	}

	return FunctionCall{
		ObjectID:            s.Field("ObjectID").MustGet().(int32),
		ArgumentFormatCodes: argumentFormatCodes,
		Arguments:           arguments,
		ResultFormatCode:    s.Field("ResultFormatCode").MustGet().(int32),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m FunctionCall) DefaultMessage() *connection.MessageFormat {
	return &functionCallDefault
}
