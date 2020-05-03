package bot

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/orivej/enlapin/bot/chatstate"
	"lukechampine.com/frand"
)

var rxEntity = regexp.MustCompile(`(^|\pZ+)[/@]\PZ*`)

func removeEntities(s string) string { return rxEntity.ReplaceAllString(s, "") }

func trim(s string) string { return strings.TrimFunc(s, unicode.IsSpace) }

func parseCard(s string) (card chatstate.Card) {
	lines := strings.Split(s, "\n")
	if len(lines) == 1 {
		topic := strings.ToLower(trim(removeEntities(lines[0])))
		if card = cardByTopic[topic]; card.Words != nil {
			return // Specific card.
		}
	}
	if parts := strings.Split(lines[0], ","); len(parts) > 1 {
		parts[0] = removeEntities(parts[0])
		lines = append(parts, lines[1:]...)
	} else {
		lines = append(strings.Fields(removeEntities(lines[0])), lines[1:]...)
	}
	for _, line := range lines {
		line = trim(line)
		if line != "" {
			card.Words = append(card.Words, line)
		}
	}
	if len(card.Words) == 0 {
		return cards[frand.Intn(len(cards))] // Random card.
	}
	return
}
