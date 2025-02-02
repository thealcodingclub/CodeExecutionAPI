package dataformating

import "strings"

func TrimLines(input string) string {
	lines := strings.Split(input, "\n")

	// Ensure we have at least 10 lines to trim
	if len(lines) <= 10 {
		return "" // Return empty if there aren't enough lines
	}

	// Slice out the lines from index 9 to the second-to-last line
	return strings.Join(lines[9:len(lines)-2], "\n")
}
