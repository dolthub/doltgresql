package messages

// SSLResponse tells the client whether SSL is supported.
type SSLResponse struct {
	SupportsSSL bool
}

// Bytes returns SSLResponse as a byte slice, ready to be returned to the client.
func (sslr SSLResponse) Bytes() []byte {
	if sslr.SupportsSSL {
		return []byte{'Y'}
	} else {
		return []byte{'N'}
	}
}
