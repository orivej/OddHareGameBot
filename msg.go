package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/orivej/OddHareGameBot/chatstate"
	"github.com/orivej/e"
	tb "gopkg.in/tucnak/telebot.v2"
)

type ctxStart struct {
	BotName string
	Private bool
}

var tmplStart = template.Must(template.New("").Parse(`
{{ define "hello" -}}
Привет! Я помощник для игры в [Зайца](https://boardgamegeek.com/boardgame/227072/chameleon)!
{{- end -}}

{{ define "Words" -}}
	{{- if .Private -}}
		Добавьте меня в группу и напишите {{""}}
	{{- else -}}
		Напишите {{""}}
        {{- end -}}
` + "`/hare слово слово слово`" + `, чтобы начать. (В оригинале 16 слов на одну тему; [примеры](http://ref.manuscriptgames.org/chameleon/index.html).) Отдельные слова можно вводить через пробел, а словосочетания через запятую или с новой строки.
{{- end -}}

{{ define "add" -}}
	{{- if .Private -}}
		Вступите {{""}}
	{{- else -}}
		[Добавьте](https://t.me/{{ .BotName }}) меня в свои контакты, вступите {{""}}
	{{- end -}}
в игру, и когда игра начнётся, я сообщу одному, что он заяц, а остальным одно из слов.
{{- end -}}

{{ define "Start" -}}
{{- template "hello" . }} {{""}}
{{- template "Words" . }} {{""}}
{{- template "add" . -}}
{{- end -}}

`))

func render(t, bot string, private bool) string {
	var buf bytes.Buffer
	err := tmplStart.ExecuteTemplate(&buf, t, ctxStart{BotName: bot, Private: private})
	e.Exit(err)
	return buf.String()
}

const msgObsolete = "Эта игра устарела, начните новую!"
const msgNoPlayers = "В игре нет игроков."
const msgYouAreHare = "Ты заяц!"

const fmtUndelievered = `%s, я не могу вам писать! <a href="https://t.me/%s">Добавьте</a> меня в свои контакты!`
const fmtPlayStarted = "Теперь каждый по очереди называет ассоциацию: %s."

type ctxChatState struct{ Players, Words []string }

var tmplChatState = template.Must(template.New("").Funcs(tmplFuncs).Parse(`
У меня {{ num (len .Words) "слов" "слово" "слова" }}. {{""}}
{{- if .Players -}}
{{ plu (len .Players) "Играет" "Играют" }} {{ joinWithAnd .Players }}. {{""}}
{{- else -}}
Ещё никто не играет. {{""}}
{{- end -}}
Присоединяйтесь!
`))

func renderChatState(players, words []string) string {
	var buf bytes.Buffer
	err := tmplChatState.Execute(&buf, ctxChatState{Players: players, Words: words})
	e.Exit(err)
	return buf.String()
}

var tmplFuncs = template.FuncMap{"num": num, "plu": plu, "joinWithAnd": joinWithAnd}

func num(x int, zero, one, two string) string {
	word := zero
	if x == 1 {
		word = one
	} else if x > 1 && x < 5 {
		word = two
	}
	return fmt.Sprintf("%d %s", x, word)
}

func plu(x int, one, two string) string {
	if x == 1 {
		return one
	}
	return two
}

func joinWithAnd(xs []string) (y string) {
	conj := ""
	for i, s := range xs {
		y += conj + s
		conj = ", "
		if i == len(xs)-2 {
			conj = " и "
		}
	}
	return
}

func joinEnumerate(xs []string) (y string) {
	if len(xs) == 0 {
		return
	}
	y = "сначала " + xs[0]
	if len(xs) > 1 {
		y += ", затем " + joinWithAnd(xs[1:])
	}
	return
}

var EscapeHTML = strings.NewReplacer(`<`, "&lt;", `>`, "&gt;", `&`, "&amp;").Replace

func PlayerHTML(cs *chatstate.ChatState, user *tb.User) string {
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`,
		user.ID, EscapeHTML(ChooseName(user, cs.Players)))
}

func PlayersHTML(cs *chatstate.ChatState, users []*tb.User) []string {
	xs := make([]string, len(users))
	for i, user := range users {
		xs[i] = PlayerHTML(cs, user)
	}
	return xs
}
