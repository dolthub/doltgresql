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
	initializeDefaultMessage(Bind{})
	addMessageHeader(Bind{})
}

// Bind represents a PostgreSQL message.
type Bind struct {
	DestinationPortal       string
	SourcePreparedStatement string
	ParameterFormatCodes    []int32
	ParameterValues         []BindParameterValue
	ResultFormatCodes       []int32
}

// BindParameterValue are parameter values for the Bind message.
type BindParameterValue struct {
	Data   []byte
	IsNull bool
}

var bindDefault = MessageFormat{
	Name: "Bind",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('B'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "DestinationPortal",
			Type: String,
			Data: "",
		},
		{
			Name: "SourcePreparedStatement",
			Type: String,
			Data: "",
		},
		{
			Name: "ParameterFormatCodes",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name: "ParameterFormatCode",
						Type: Int16,
						Data: int32(0),
					},
				},
			},
		},
		{
			Name: "ParameterValues",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name:  "ParameterLength",
						Type:  Int32,
						Flags: ByteCount,
						Data:  int32(0),
					},
					{
						Name: "ParameterValue",
						Type: ByteN,
						Data: []byte{},
					},
				},
			},
		},
		{
			Name: "ResultFormatCodes",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
				{
					{
						Name: "ResultFormatCode",
						Type: Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ Message = Bind{}

// encode implements the interface Message.
func (m Bind) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
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

// decode implements the interface Message.
func (m Bind) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
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

// defaultMessage implements the interface Message.
func (m Bind) defaultMessage() *MessageFormat {
	return &bindDefault
}
