package chatstate

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

const Lifetime = 6 * time.Hour // How long is the ChatState usable after Last.Unixtime.
const Locktime = time.Minute   // How long can a handler exclusively use a ChatState.

type Card struct {
	Topic string
	Words []string
}

type ChatState struct {
	Last          *tb.Message // Last /play response.
	LastBeginTime time.Time   // Last Begin (Play or Undelivered) response time.
	Players       []*tb.User
	Card          Card
}

func (cs *ChatState) Reset() {
	*cs = ChatState{}
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
