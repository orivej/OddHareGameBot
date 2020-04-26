package main

import "sync"

type LocalChatStateMap struct{ sync.Map }
type LocalChatStateMapItem struct {
	sync.Mutex
	ChatState
}

func NewLocalChatStateMap() ChatStateMap {
	return &LocalChatStateMap{}
}

func (m *LocalChatStateMap) Get(chatID int64) (*ChatState, func()) {
	item := &LocalChatStateMapItem{}
	x, _ := m.LoadOrStore(chatID, item)
	item = x.(*LocalChatStateMapItem)
	item.Lock()
	return &item.ChatState, item.Unlock
}
