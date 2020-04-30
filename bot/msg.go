package bot

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/orivej/e"
	"github.com/orivej/enlapin/bot/chatstate"
	tb "gopkg.in/tucnak/telebot.v2"
)

const msgObsolete = "Эта игра устарела, начните новую! /hare"
const msgNoPlayers = "В игре нет игроков."
const msgYouAreHare = "Ты заяц!"
const fmtPlayStarted = "Теперь каждый по очереди называет ассоциацию: %s."
const msgRules = `
Участники, трое или больше, получают набор слов: 16 слов на одну тему. Игра случайно выбирает слово из набора, зайца из игроков и очерёдность хода. Игра сообщает зайцу, что он заяц, а остальным загаданное слово. Каждый по очереди объявляет одно слово по ассоциации с тайным; заяц в том числе, даже если ходит первым. Все обсуждают, кто же заяц, и голосуют — при ничьей первый игрок решает, кто из лидеров голосования заяц. Заяц выигрывает, если выбрали не его или если он правильно назовёт тайное слово с первой, а при игре втроём со второй попытки.

В игре на счёт до пяти очков непойманный заяц получает два очка, осведомлённый одно, иначе остальные по два. Настольная версия: https://www.mosigra.ru/Face/Show/zayac/ Разбор игры: https://habr.com/ru/company/mosigra/blog/456844/
`

var tmplFuncs = template.FuncMap{"num": num, "plu": plu, "joinWithAnd": joinWithAnd}

var tmpl = template.Must(template.New("").Funcs(tmplFuncs).Parse(`
{{ define "hello" -}}
Привет! Я помощник для игры в [Зайца](https://boardgamegeek.com/boardgame/227072/chameleon)! /rules — правила.
{{- end -}}

{{ define "Words" -}}
	{{- if .Private -}}
		Добавь меня в группу и напиши {{""}}
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

{{ define "ChatState" -}}
У меня {{ num (len .Words) "слов" "слово" "слова" }}. {{""}}
{{- if .Players -}}
{{ plu (len .Players) "Играет" "Играют" }} {{ joinWithAnd .Players }}. {{""}}
{{- else -}}
Ещё никто не играет. {{""}}
{{- end -}}
Присоединяйтесь!
{{- end -}}

{{ define "Undelivered" -}}
{{ joinWithAnd .Players }}, я не могу
{{- if eq 1 (len .Players) }} тебе {{ else }} вам {{ end -}}
писать! <a href="https://t.me/{{ .BotName }}">Добавь
{{- if ne 1 (len .Players) }}те{{ end -}}
</a> меня в свои контакты!
{{- end -}}
`))

func render(name string, ctx interface{}) string {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, ctx)
	e.Exit(err)
	return buf.String()
}

func renderHelp(t, bot string, private bool) string {
	return render(t, struct {
		BotName string
		Private bool
	}{bot, private})
}

func renderChatState(players, words []string) string {
	return render("ChatState", struct{ Players, Words []string }{players, words})
}

func renderUndelievered(players []string, bot string) string {
	return render("Undelivered", struct {
		Players []string
		BotName string
	}{players, bot})
}

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
