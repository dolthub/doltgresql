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
	initializeDefaultMessage(CancelRequest{})
}

// CancelRequest represents a PostgreSQL message.
type CancelRequest struct {
	ProcessID int32
	SecretKey int32
}

var cancelRequestDefault = MessageFormat{
	Name: "CancelRequest",
	Fields: FieldGroup{
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "RequestCode",
			Type: Int32,
			Data: int32(80877102),
		},
		{
			Name: "ProcessID",
			Type: Int32,
			Data: int32(0),
		},
		{
			Name: "SecretKey",
			Type: Int32,
			Data: int32(0),
		},
	},
}

var _ Message = CancelRequest{}

// encode implements the interface Message.
func (m CancelRequest) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("ProcessID").MustWrite(m.ProcessID)
	outputMessage.Field("SecretKey").MustWrite(m.SecretKey)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m CancelRequest) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return CancelRequest{
		ProcessID: s.Field("ProcessID").MustGet().(int32),
		SecretKey: s.Field("SecretKey").MustGet().(int32),
	}, nil
}

// defaultMessage implements the interface Message.
func (m CancelRequest) defaultMessage() *MessageFormat {
	return &cancelRequestDefault
}
