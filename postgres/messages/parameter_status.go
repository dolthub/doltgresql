package messages

import "bytes"

// ParameterStatus reports various parameters to the client.
type ParameterStatus struct {
	Name  string
	Value string
}

// Bytes returns ParameterStatus as a byte slice, ready to be returned to the client.
func (ps ParameterStatus) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('S')          // Message Type
	WriteNumber(&buf, int32(0)) // Message length, will be corrected later
	buf.WriteString(ps.Name)
	buf.WriteByte(0) // Trailing NULL character, denoting the end of the string
	buf.WriteString(ps.Value)
	buf.WriteByte(0) // Trailing NULL character, denoting the end of the string
	return WriteLength(buf.Bytes())
}
