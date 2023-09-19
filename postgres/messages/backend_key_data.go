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
	initializeDefaultMessage(BackendKeyData{})
	addMessageHeader(BackendKeyData{})
}

// BackendKeyData provides the client with information about the server.
type BackendKeyData struct {
	ProcessID int32
	SecretKey int32
}

var backendKeyDataDefault = Message{
	Name: "BackendKeyData",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('K'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(12),
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

var _ MessageType = BackendKeyData{}

// encode implements the interface MessageType.
func (m BackendKeyData) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("ProcessID").MustWrite(m.ProcessID)
	outputMessage.Field("SecretKey").MustWrite(m.SecretKey)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m BackendKeyData) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return BackendKeyData{
		ProcessID: s.Field("ProcessID").MustGet().(int32),
		SecretKey: s.Field("SecretKey").MustGet().(int32),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m BackendKeyData) defaultMessage() *Message {
	return &backendKeyDataDefault
}
