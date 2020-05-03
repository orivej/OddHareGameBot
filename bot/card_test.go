package bot

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/orivej/enlapin/bot/chatstate"
)

func TestSplitCards(t *testing.T) {
	for s, cards := range map[string][]chatstate.Card{
		"a: b":            {{"a", []string{"b"}}},
		"\na: b\n":        {{"a", []string{"b"}}},
		"a: b c, d":       {{"a", []string{"b c", "d"}}},
		"a: b, c\nd: e f": {{"a", []string{"b", "c"}}, {"d", []string{"e f"}}},
	} {
		assert.Equal(t, cards, splitCards(s))
	}
}

func TestCards(t *testing.T) {
	assert.NotEmpty(t, cards)
	assert.Len(t, cardByTopic, len(cards))
	for _, card := range cards {
		assert.NotContains(t, card.Topic, ",")
		assert.InDelta(t, 0, len(encodeTopic(card.Topic)), 64.0, card.Topic)
		for _, word := range card.Words {
			assert.True(t, word[0] != ' ', word)
			assert.True(t, word[len(word)-1] != ' ', word)
		}
	}
}
