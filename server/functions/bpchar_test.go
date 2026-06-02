package functions

import (
	"strings"
	"testing"
)

// BenchmarkTruncateString-14    	   25954	     46215 ns/op
func BenchmarkTruncateString(b *testing.B) {
	str := strings.Repeat("a", 65535)
	for i := 0; i < b.N; i++ {
		_, _ = truncateString(str, 65535)
	}
}
