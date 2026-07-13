package templateapp

import "strings"

const MaxEchoMessageLen = 500

func NormalizeEchoMessage(s string) string {
	return strings.TrimSpace(s)
}
