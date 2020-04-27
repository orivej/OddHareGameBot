package chatstate

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

const Lifetime = 6 * time.Hour // How long is the ChatState usable after Last.Unixtime.
const Locktime = time.Minute   // How long can a handler exclusively use a ChatState.

type ChatState struct {
	Last    *tb.Message
	Players []*tb.User
	Words   []string
}

func (cs *ChatState) Reset() {
	cs.Last, cs.Words, cs.Players = nil, nil, nil
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

func (cs *ChatState) RemovePlayer(user *tb.User) bool {
	for i, player := range cs.Players {
		if user.ID == player.ID {
			copy(cs.Players[i:], cs.Players[i+1:])
			cs.Players = cs.Players[:len(cs.Players)-1]
			return true
		}
	}
	return false
}
