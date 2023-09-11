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
