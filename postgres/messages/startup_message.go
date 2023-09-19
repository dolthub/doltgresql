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
	"encoding/binary"
	"fmt"
)

// StartupMessage is returned by the client upon connecting to the server, providing details about the client.
type StartupMessage struct {
	ProtocolMajorVersion int16
	ProtocolMinorVersion int16
	Parameters           map[string]string
}

// ReadStartupMessage returns the StartupMessage from the buffer.
func ReadStartupMessage(buf []byte) (StartupMessage, error) {
	if len(buf) < 4 {
		return StartupMessage{}, fmt.Errorf("invalid StartupMessage")
	}
	messageLength := int32(binary.BigEndian.Uint32(buf))
	protocolMajorVersion := int16(binary.BigEndian.Uint16(buf[4:]))
	protocolMinorVersion := int16(binary.BigEndian.Uint16(buf[6:]))
	// Set the buffer to the stated length and skip the length and version
	buf = buf[8:messageLength]
	parameters := make(map[string]string)
	for len(buf) > 0 {
		var name string
		var value string
		for i, b := range buf {
			if b == 0 {
				name = string(buf[:i])
				buf = buf[i+1:]
				break
			}
		}
		for i, b := range buf {
			if b == 0 {
				value = string(buf[:i])
				buf = buf[i+1:]
				break
			}
		}
		if len(name) > 0 && len(value) > 0 {
			parameters[name] = value
		}
	}
	return StartupMessage{
		ProtocolMajorVersion: protocolMajorVersion,
		ProtocolMinorVersion: protocolMinorVersion,
		Parameters:           parameters,
	}, nil
}
