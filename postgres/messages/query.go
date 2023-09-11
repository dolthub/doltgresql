package messages

import (
	"encoding/binary"
)

// ReadQuery returns the query from the given buffer. Assumes that the buffer contains a serialized form of a Query
// message.
func ReadQuery(buf []byte) (string, bool) {
	if len(buf) < 5 {
		return "", false
	}
	if buf[0] != 'Q' {
		return "", false
	}
	queryLength := int32(binary.BigEndian.Uint32(buf[1:]))
	if queryLength <= 5 {
		// A query of length 5 or less is empty
		return "", true
	}
	// The length includes the length bytes, along with the NULL terminator. It does not include the message identifier
	// though, so it cancels out the NULL terminator and allows us to use the length as-is.
	return string(buf[5:queryLength]), true
}
