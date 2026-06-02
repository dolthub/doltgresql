package functions

import (
	"strings"
	"testing"
)

func BenchmarkTruncateString(b *testing.B) {
	str := strings.Repeat("a", 65535)
	for i := 0; i < b.N; i++ {
		_, _ = truncateString(str, 65535)
	}
}
