package queryutil

import "strings"

func Clean(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
