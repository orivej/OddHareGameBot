package bot

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestRemoveEntities(t *testing.T) {
	for line, filtered := range map[string]string{
		"":                     "",
		"/hare @a   @b c  d e": " c  d e",
	} {
		assert.Equal(t, filtered, removeEntities(line))
	}
}

func TestParseWords(t *testing.T) {
	for query, words := range map[string][]string{
		"": nil,
		"/hare  @a   @b c  d \n  e f \n \n g,  h ": {"c", "d", "e f", "g,  h"},
	} {
		assert.Equal(t, words, parseWords(query))
	}
}
