package main

import (
	"regexp"
	"strings"
	"unicode"
)

var rxEntity = regexp.MustCompile(`(^|\pZ+)[/@]\PZ*`)

func removeEntities(s string) string { return rxEntity.ReplaceAllString(s, "") }

func trim(s string) string { return strings.TrimFunc(s, unicode.IsSpace) }

func parseWords(s string) (words []string) {
	lines := strings.Split(s, "\n")
	if parts := strings.Split(lines[0], ","); len(parts) > 1 {
		parts[0] = removeEntities(parts[0])
		lines = append(parts, lines[1:]...)
	} else {
		lines = append(strings.Fields(removeEntities(lines[0])), lines[1:]...)
	}
	for _, line := range lines {
		line = trim(line)
		if line != "" {
			words = append(words, line)
		}
	}
	return
}
