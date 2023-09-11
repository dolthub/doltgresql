package messages

import "bytes"

// BackendKeyData provides the client with information about the server.
type BackendKeyData struct {
	ProcessID int32
	SecretKey int32
}

// Bytes returns BackendKeyData as a byte slice, ready to be returned to the client.
func (bkd BackendKeyData) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('K')           // Message Type
	WriteNumber(&buf, int32(12)) // Message Length
	WriteNumber(&buf, bkd.ProcessID)
	WriteNumber(&buf, bkd.SecretKey)
	return buf.Bytes()
}
