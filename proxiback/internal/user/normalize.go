package user

import "strings"

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeDisplayName(displayName string) string {
	return strings.TrimSpace(displayName)
}
