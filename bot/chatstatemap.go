package bot

import (
	"sync"

	"github.com/orivej/enlapin/bot/chatstate"
)

type LocalChatStateMap struct{ sync.Map }
type LocalChatStateMapItem struct {
	sync.Mutex
	chatstate.ChatState
}

func NewLocalChatStateMap() *LocalChatStateMap {
	return &LocalChatStateMap{}
}

func (m *LocalChatStateMap) Get(chatID int64) (*chatstate.ChatState, func()) {
	item := &LocalChatStateMapItem{}
	x, _ := m.LoadOrStore(chatID, item)
	item = x.(*LocalChatStateMapItem)
	item.Lock()
	return &item.ChatState, item.Unlock
}
