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

import "fmt"

func init() {
	initializeDefaultMessage(SSLResponse{})
}

// SSLResponse tells the client whether SSL is supported.
type SSLResponse struct {
	SupportsSSL bool
}

var sslResponseDefault = MessageFormat{
	Name: "SSLResponse",
	Fields: FieldGroup{
		{
			Name: "Supported",
			Type: Byte1,
			Data: int32(0),
		},
	},
}

var _ Message = SSLResponse{}

// encode implements the interface Message.
func (m SSLResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	if m.SupportsSSL {
		outputMessage.Field("Supported").MustWrite('Y')
	} else {
		outputMessage.Field("Supported").MustWrite('N')
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m SSLResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	var supported bool
	supportedInt := s.Field("Supported").MustGet().(int32)
	if supportedInt == 'Y' {
		supported = true
	} else if supportedInt == 'N' {
		supported = false
	} else {
		return nil, fmt.Errorf("Unexpected supported value in SSLResponse message: %d", supportedInt)
	}
	return SSLResponse{
		SupportsSSL: supported,
	}, nil
}

// defaultMessage implements the interface Message.
func (m SSLResponse) defaultMessage() *MessageFormat {
	return &sslResponseDefault
}
