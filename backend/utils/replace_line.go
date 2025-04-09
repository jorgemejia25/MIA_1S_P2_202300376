package utils

import "strings"

func ReplaceLine(fileData string, lineIndex int, newLine string) string {
	lines := strings.Split(fileData, "\n")
	if lineIndex >= 0 && lineIndex < len(lines) {
		lines[lineIndex] = newLine
	}
	return strings.Join(lines, "\n")
}
