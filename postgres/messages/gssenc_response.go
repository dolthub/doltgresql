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

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(GSSENCResponse{})
}

// GSSENCResponse tells the client whether GSSAPI is supported. This is not a "true" PostgreSQL message, as it is not
// defined by the documentation as a message. However, clients expect to receive a response, so that response is encoded
// as a message.
type GSSENCResponse struct {
	SupportsGSSAPI bool
}

var gssENCResponseDefault = connection.MessageFormat{
	Name: "GSSENCResponse",
	Fields: connection.FieldGroup{
		{
			Name: "Supported",
			Type: connection.Byte1,
			Data: int32(0),
		},
	},
}

var _ connection.Message = GSSENCResponse{}

// Encode implements the interface connection.Message.
func (m GSSENCResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	if m.SupportsGSSAPI {
		outputMessage.Field("Supported").MustWrite('G')
	} else {
		outputMessage.Field("Supported").MustWrite('N')
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m GSSENCResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	var supported bool
	supportedInt := s.Field("Supported").MustGet().(int32)
	if supportedInt == 'G' {
		supported = true
	} else if supportedInt == 'N' {
		supported = false
	} else {
		return nil, fmt.Errorf("Unexpected supported value in SSLResponse message: %d", supportedInt)
	}
	return GSSENCResponse{
		SupportsGSSAPI: supported,
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m GSSENCResponse) DefaultMessage() *connection.MessageFormat {
	return &gssENCResponseDefault
}
