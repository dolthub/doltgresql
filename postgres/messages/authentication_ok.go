package messages

// AuthenticationOk tells the client that authentication was successful.
type AuthenticationOk struct{}

// Bytes returns AuthenticationOk as a byte slice, ready to be returned to the client.
func (aok AuthenticationOk) Bytes() []byte {
	return []byte{
		'R',        // Message Type
		0, 0, 0, 8, // Message Length
		0, 0, 0, 0, // Padding
	}
}
