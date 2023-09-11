package messages

// ReadTerminate returns whether the buffer represents a Terminate message.
func ReadTerminate(buf []byte) bool {
	if len(buf) < 5 {
		return false
	}
	return buf[0] == 'X' && buf[4] == 4
}
