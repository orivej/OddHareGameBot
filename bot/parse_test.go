package bot

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestRemoveEntities(t *testing.T) {
	for line, filtered := range map[string]string{
		"":                     "",
		"/play @a   @b c  d e": " c  d e",
	} {
		assert.Equal(t, filtered, removeEntities(line))
	}
}

func TestParseCard(t *testing.T) {
	assert.NotEmpty(t, parseCard("").Words)
	assert.NotEmpty(t, parseCard("ГОРОДА").Words)
	for query, words := range map[string][]string{
		"города": cardByTopic["города"].Words,
		"/play  @a   @b c  d \n  e f \n \n g,  h ": {"c", "d", "e f", "g,  h"},
	} {
		assert.Equal(t, words, parseCard(query).Words)
	}
}
