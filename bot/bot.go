package bot

import (
	"fmt"
	"time"

	"github.com/orivej/OddHareGameBot/bot/chatstate"
	"github.com/orivej/OddHareGameBot/bot/ddb"
	"github.com/orivej/e"
	tb "gopkg.in/tucnak/telebot.v2"
	"lukechampine.com/frand"
)

type ChatStateMap interface {
	Get(chatID int64) (*chatstate.ChatState, func())
}

type Bot struct {
	*tb.Bot
	Name  string
	Chats ChatStateMap

	BtnJoin  tb.InlineButton
	BtnLeave tb.InlineButton
	BtnPlay  tb.InlineButton
}

func NewBot(b *tb.Bot, name string, local bool, table string) *Bot {
	bot := &Bot{Bot: b, Name: name}
	if local {
		bot.Chats = NewLocalChatStateMap()
	} else {
		bot.Chats = ddb.NewDDBChatStateMap(table)
	}
	return bot
}

func (b *Bot) Setup() {
	b.BtnJoin = tb.InlineButton{Unique: "join", Text: "Вступить"}
	b.BtnLeave = tb.InlineButton{Unique: "leave", Text: "Выйти"}
	b.BtnPlay = tb.InlineButton{Unique: "play", Text: "Играть"}
	b.Handle("/start", b.OnStart)
	b.Handle(tb.OnAddedToGroup, b.OnStart)
	b.Handle("/hare", b.OnHare)
	b.Handle(&b.BtnJoin, b.OnBtnJoin)
	b.Handle(&b.BtnLeave, b.OnBtnLeave)
	b.Handle(&b.BtnPlay, b.OnBtnPlay)
}

func (b *Bot) Post(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, bool) {
	m, err := b.Send(to, what, options...)
	e.Print(err)
	return m, err == nil
}

func (b *Bot) OnStart(m *tb.Message) {
	msg := render("Start", b.Name, m.Chat.Type == tb.ChatPrivate)
	b.Post(m.Chat, msg, tb.ModeMarkdown, tb.NoPreview)
}

func (b *Bot) OnHare(m *tb.Message) {
	words := parseWords(m.Text)
	if len(words) == 0 {
		msg := render("Words", b.Name, m.Chat.Type == tb.ChatPrivate)
		b.Post(m.Chat, msg, tb.ModeMarkdown, tb.NoPreview)
		return
	}
	cs, unlock := b.Chats.Get(m.Chat.ID)
	defer unlock()
	if cs.Last != nil {
		// Delete keyboard.
		_, err := b.EditReplyMarkup(cs.Last, &tb.ReplyMarkup{})
		e.Print(err)
	}
	cs.Reset()
	cs.Words = words
	b.PostChatState(m.Chat, cs)
}

func (b *Bot) ChatStateMessage(cs *chatstate.ChatState) (what string, options []interface{}) {
	msg := renderChatState(PlayersHTML(cs, cs.Players), cs.Words)
	return msg, []interface{}{tb.ModeHTML, tb.NoPreview, &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{{b.BtnJoin, b.BtnLeave, b.BtnPlay}},
	}}
}

func (b *Bot) PostChatState(chat *tb.Chat, cs *chatstate.ChatState) {
	msg, opts := b.ChatStateMessage(cs)
	if resp, ok := b.Post(chat, msg, opts...); ok {
		cs.Last = resp
	}
}

func (b *Bot) UpdateChatState(cs *chatstate.ChatState) {
	if cs.Last == nil {
		return
	}
	msg, opts := b.ChatStateMessage(cs)
	_, err := b.Edit(cs.Last, msg, opts...)
	e.Print(err)
}

func (b *Bot) replyObselete(m *tb.Message) bool {
	if time.Now().Before(m.Time().Add(chatstate.Lifetime)) {
		return false
	}
	_, err := b.Reply(m, msgObsolete)
	e.Print(err)
	_, err = b.EditReplyMarkup(m, &tb.ReplyMarkup{}) // Delete old keyboard.
	e.Print(err)
	return true
}

func (b *Bot) OnBtnJoin(c *tb.Callback) {
	m := c.Message
	if b.replyObselete(m) {
		return
	}
	cs, unlock := b.Chats.Get(m.Chat.ID)
	defer unlock()
	if cs.AddPlayer(c.Sender) {
		b.UpdateChatState(cs)
	}
}
func (b *Bot) OnBtnLeave(c *tb.Callback) {
	m := c.Message
	if b.replyObselete(m) {
		return
	}
	cs, unlock := b.Chats.Get(m.Chat.ID)
	defer unlock()
	if cs.RemovePlayer(c.Sender) {
		b.UpdateChatState(cs)
	}
}

func (b *Bot) OnBtnPlay(c *tb.Callback) {
	m := c.Message
	if b.replyObselete(m) {
		return
	}
	cs, unlock := b.Chats.Get(m.Chat.ID)
	defer unlock()
	if len(cs.Players) == 0 {
		_, err := b.Reply(m, msgNoPlayers)
		e.Print(err)
		return
	}
	frand.Shuffle(len(cs.Players), func(i, j int) {
		cs.Players[i], cs.Players[j] = cs.Players[j], cs.Players[i]
	})
	hare := cs.Players[frand.Intn(len(cs.Players))]
	word := cs.Words[frand.Intn(len(cs.Words))]
	var failed []*tb.User
	for _, player := range cs.Players {
		msg := word
		if player == hare {
			msg = msgYouAreHare
		}
		_, err := b.Send(player, msg)
		if err != nil {
			if !(err == tb.ErrNotStartedByUser || err == tb.ErrBlockedByUser) {
				e.Print(err)
			}
			failed = append(failed, player)
		}
	}
	if len(failed) > 0 {
		msg := fmt.Sprintf(fmtUndelievered, joinWithAnd(PlayersHTML(cs, failed)), b.Name)
		b.Post(m.Chat, msg, tb.ModeHTML, tb.NoPreview)
		return
	}
	msg := fmt.Sprintf(fmtPlayStarted, joinEnumerate(PlayersHTML(cs, cs.Players)))
	b.Post(m.Chat, msg, tb.ModeHTML, tb.NoPreview)
}
