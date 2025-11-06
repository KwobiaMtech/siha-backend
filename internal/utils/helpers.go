package utils

import "strings"

func SplitName(fullName string) [2]string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return [2]string{"", ""}
	}
	if len(parts) == 1 {
		return [2]string{parts[0], ""}
	}
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return [2]string{firstName, lastName}
}
