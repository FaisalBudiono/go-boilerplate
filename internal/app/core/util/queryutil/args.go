package queryutil

import (
	"fmt"
	"strings"
)

func ArgsPlaceholder(argsTotal, argNoBefore int) string {
	args := make([]string, argsTotal)
	for i := range argsTotal {
		argNo := i + argNoBefore + 1

		args[i] = fmt.Sprintf("$%d", argNo)
	}

	return strings.Join(args, ", ")
}
