package queryutil

import "strings"

func Clean(s string) string {
	// noNL := strings.ReplaceAll(s, "\n", "")
	return strings.Join(strings.Fields(s), " ")
}
