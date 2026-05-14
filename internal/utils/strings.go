package utils

import "strings"

func NormalizeKey(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
