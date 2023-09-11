package messages

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/exp/constraints"
)

// WriteLength writes the length of the message into the byte slice. Modifies the given byte slice, while also
// returning the same slice. Assumes that the first byte is the message identifier, while the next 4 bytes are
// the length.
func WriteLength(b []byte) []byte {
	// We never include the message identifier in the length.
	// Technically, the length field is an int32, however we'll assume that our return values will be under 2GB for now.
	length := uint32(len(b) - 1)
	binary.BigEndian.PutUint32(b[1:], length)
	return b
}

// WriteNumber writes the given number to the buffer.
func WriteNumber[T constraints.Integer | constraints.Float](buf *bytes.Buffer, num T) {
	_ = binary.Write(buf, binary.BigEndian, num)
}
