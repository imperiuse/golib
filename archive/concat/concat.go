package concat

import "strings"

func Strings(args ...string) string {
	var strBuilder strings.Builder
	for _, arg := range args {
		strBuilder.WriteString(arg)
	}
	return strBuilder.String()
}
