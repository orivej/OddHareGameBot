package bot

import (
	"time"

	"github.com/orivej/e"
	"github.com/orivej/enlapin/bot/chatstate"
	"github.com/orivej/enlapin/bot/ddb"
	tb "gopkg.in/tucnak/telebot.v2"
	"lukechampine.com/frand"
)

type ChatStateMap interface {
	Get(chatID int64) (*chatstate.ChatState, func())
}

type Bot struct {
	*tb.Bot
	Chats ChatStateMap

	BtnJoin  tb.InlineButton
	BtnLeave tb.InlineButton
	BtnPlay  tb.InlineButton
}

func NewBot(b *tb.Bot, local bool, table string) *Bot {
	bot := &Bot{Bot: b}
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
	b.Handle("/rules", b.OnRules)
	b.Handle("/topics", b.OnTopics)
	b.Handle("/about", b.OnAbout)
	b.Handle("/aboutname", b.OnAboutName)
	b.Handle("/aboutpic", b.OnAboutPic)
	b.Handle("/aboutid", b.OnAboutID)
	b.Handle("/hare", b.OnHare)
	b.Handle(&b.BtnJoin, b.OnBtnJoin)
	b.Handle(&b.BtnLeave, b.OnBtnLeave)
	b.Handle(&b.BtnPlay, b.OnBtnPlay)
}

func (b *Bot) Post(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, bool) {
	options = append(options, tb.NoPreview)
	m, err := b.Send(to, what, options...)
	e.Print(err)
	return m, err == nil
}

func (b *Bot) OnStart(m *tb.Message) {
	if m.Payload == "startgroup" {
		return
	}
	if topic := decodeTopic(m.Payload); topic != "" {
		m.Text = topic
		b.OnHare(m)
		return
	}
	msg := renderHelp("Start", b.Me.Username, m.Private())
	b.Post(m.Chat, msg, tb.ModeMarkdown)
}

func (b *Bot) OnRules(m *tb.Message) { b.Post(m.Chat, msgRules, tb.ModeMarkdown, tb.NoPreview) }
func (b *Bot) OnTopics(m *tb.Message) {
	b.Post(m.Chat, renderTopics(b.Me.Username, m.Private()), tb.ModeHTML)
}
func (b *Bot) OnAbout(m *tb.Message)     { b.Post(m.Chat, msgAbout, tb.ModeMarkdown) }
func (b *Bot) OnAboutName(m *tb.Message) { b.Post(m.Chat, msgAboutName, tb.ModeMarkdown) }
func (b *Bot) OnAboutPic(m *tb.Message)  { b.Post(m.Chat, msgAboutPic, tb.ModeMarkdown) }
func (b *Bot) OnAboutID(m *tb.Message)   { b.Post(m.Chat, msgAboutID, tb.ModeMarkdown) }

func (b *Bot) OnHare(m *tb.Message) {
	cs, unlock := b.Chats.Get(m.Chat.ID)
	defer unlock()
	if cs.Last != nil {
		// Delete keyboard.
		_, err := b.EditReplyMarkup(cs.Last, &tb.ReplyMarkup{})
		e.Print(err)
	}
	cs.Reset()
	cs.Card = parseCard(m.Text)
	b.PostChatState(m.Chat, cs)
}

func (b *Bot) ChatStateMessage(cs *chatstate.ChatState) (what string, options []interface{}) {
	msg := renderChatState(cs)
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

func (b *Bot) checkObselete(m *tb.Message) bool {
	if time.Now().Before(m.Time().Add(chatstate.Lifetime)) {
		return false
	}
	b.replyObselete(m)
	return true
}

func (b *Bot) replyObselete(m *tb.Message) {
	_, err := b.Reply(m, msgObsolete)
	e.Print(err)
	_, err = b.EditReplyMarkup(m, &tb.ReplyMarkup{}) // Delete old keyboard.
	e.Print(err)
}

func (b *Bot) chatState(m *tb.Message) (cs *chatstate.ChatState, unlock func()) {
	if !b.checkObselete(m) {
		cs, unlock = b.Chats.Get(m.Chat.ID)
		if cs == nil || cs.Last == nil || cs.Last.ID != m.ID {
			b.replyObselete(m)
			unlock()
			return nil, nil
		}
	}
	return
}

func (b *Bot) OnBtnJoin(c *tb.Callback) {
	m := c.Message
	cs, unlock := b.chatState(m)
	if cs == nil {
		return
	}
	defer unlock()
	if cs.AddPlayer(c.Sender) {
		b.UpdateChatState(cs)
	}
}
func (b *Bot) OnBtnLeave(c *tb.Callback) {
	m := c.Message
	cs, unlock := b.chatState(m)
	if cs == nil {
		return
	}
	defer unlock()
	if cs.RemovePlayer(c.Sender) {
		b.UpdateChatState(cs)
	}
}

func (b *Bot) OnBtnPlay(c *tb.Callback) {
	m := c.Message
	cs, unlock := b.chatState(m)
	if cs == nil {
		return
	}
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
	word := cs.Card.Words[frand.Intn(len(cs.Card.Words))]
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
		msg := renderUndelievered(PlayersHTML(cs, failed), b.Me.Username)
		b.Post(m.Chat, msg, tb.ModeHTML)
		return
	}
	msg := renderPlay(PlayersHTML(cs, cs.Players), b.Me.Username, m.Private())
	b.Post(m.Chat, msg, tb.ModeHTML)
}
