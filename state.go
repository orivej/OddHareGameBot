package main

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type ChatState struct {
	Last    *tb.Message
	Players []*tb.User
	Words   []string
}

func (cs *ChatState) Reset() {
	cs.Last, cs.Words, cs.Players = nil, nil, nil
}

func (cs *ChatState) Describe() string {
	return renderChatState(cs.PlayersHTML(cs.Players), cs.Words)
}

func (cs *ChatState) AddPlayer(user *tb.User) bool {
	for _, player := range cs.Players {
		if user.ID == player.ID {
			return false
		}
	}
	cs.Players = append(cs.Players, user)
	return true
}

var EscapeHTML = strings.NewReplacer(`<`, "&lt;", `>`, "&gt;", `&`, "&amp;").Replace

func (cs *ChatState) PlayerHTML(user *tb.User) string {
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`,
		user.ID, EscapeHTML(ChooseName(user, cs.Players)))
}

func (cs *ChatState) PlayersHTML(users []*tb.User) []string {
	xs := make([]string, len(users))
	for i, user := range users {
		xs[i] = cs.PlayerHTML(user)
	}
	return xs
}
