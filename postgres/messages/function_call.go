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

func init() {
	initializeDefaultMessage(FunctionCall{})
	addMessageHeader(FunctionCall{})
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

var functionCallDefault = MessageFormat{
	Name: "FunctionCall",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('F'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "ObjectID",
			Type: Int32,
			Data: int32(0),
		},
		{
			Name: "ArgumentFormatCodes",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name: "ArgumentFormatCode",
						Type: Int16,
						Data: int32(0),
					},
				},
			},
		},
		{
			Name: "Arguments",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name:  "ArgumentLength",
						Type:  Int32,
						Flags: ByteCount,
						Data:  int32(0),
					},
					{
						Name: "ArgumentValue",
						Type: ByteN,
						Data: []byte{},
					},
				},
			},
		},
		{
			Name: "ResultFormatCode",
			Type: Int16,
			Data: int32(0),
		},
	},
}

var _ Message = FunctionCall{}

// encode implements the interface Message.
func (m FunctionCall) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
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

// decode implements the interface Message.
func (m FunctionCall) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
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

// defaultMessage implements the interface Message.
func (m FunctionCall) defaultMessage() *MessageFormat {
	return &functionCallDefault
}
