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
Участники, трое или больше, вводят набор слов: 16 слов на одну тему. Игра случайно выбирает слово из набора, зайца из игроков и очерёдность хода. Игра сообщает зайцу, что он заяц, а остальным загаданное слово. Каждый по очереди объявляет одно слово по ассоциации с тайным; заяц в том числе, даже если ходит первым. Все обсуждают, кто же заяц, и голосуют — при ничьей первый игрок решает, кто из лидеров голосования заяц. Заяц выигрывает, если выбрали не его или если он правильно назовёт тайное слово с первой, а при игре втроём со второй попытки.

В игре на счёт до пяти очков непойманный заяц получает два очка, осведомлённый одно, иначе остальные по два. Настольная версия: https://www.mosigra.ru/Face/Show/zayac/ Разбор игры: https://habr.com/ru/company/mosigra/blog/456844/
`
const msgAbout = `
Это [проект с открытым кодом](https://github.com/orivej/enlapin), его можно посмотреть и изменить!
/aboutname — Кто такой Люциус Кларк?
/aboutpic — Кто изображён на моём аватаре?
/aboutid — Откуда пошло выражение «ехать зайцем»?
`
const msgAboutName = `
В мастерской Люциуса Кларка было полным-полно кукол: куклы-дамы и куклы-пупсы, куклы, чьи глаза открывались и закрывались, и куклы с нарисованными глазами, а ещё куклы-королевы и куклы в матросских костюмчиках.
– Что ты такое? – спросила одна высоким голоском, когда Эдварда посадили рядом с ней на полку.
– Я кролик, – ответил Эдвард.
Кукла припискнула.
– Тогда ты попал не по адресу, – сказала она. – Здесь продают кукол, а не кроликов.
Эдвард промолчал.
– Кшш! – не унималась соседка.
– Я бы с радостью, – сказал Эдвард, – но совершенно очевидно, что сам я отсюда не слезу.

Кейт ДиКамилло. Удивительное путешествие кролика Эдварда.
`
const msgAboutPic = `
На [карикатуре](https://digi.ub.uni-heidelberg.de/diglit/caricatures1870_1871bd4/0045/image) Фостена Бетбедера (из [коллекции музея Карнавале](http://parismuseescollections.paris.fr/fr/musee-carnavalet/oeuvres/le-lievre-0)) заяц — генерал и губернатор осаждаемого в 1870-м немецкими войсками Парижа Луи Жюль Трошю — держит пропуск, выписанный ему канцлером Бисмарком в Версале, ставшем штабом прусского командования и [местом провозглашения](https://commons.wikimedia.org/wiki/File:Wernerprokla.jpg) единой Германской империи — грандиозного достижения умелого дипломата. (Вот был бы Заяц!)

Насколько мне известно, Отто фон Бисмарк не занимался пропусками — карикатура основана на другом: дело в том, что из окружённого Парижа было не выехать без разрешения Трошю. Об этом по-британски [комично пишет](https://archive.org/details/cu31924028286981/page/n135/mode/2up/search/laisser-passer) Эрнест Визетелли (сын переводчика и издателя Золя) в своих Днях приключений: падение Франции, 1870-71. А когда Виктор Гюго верулся на время осады, чтобы поддержать сограждан (сборник этого периода будет назван Грозный год и посвящён Парижу, столице народов), ему тоже пришлось обратиться к губернатору:

Генералу Трошю
Париж, 25 сентября 1870
Генерал, старик сам по себе — ничто. Пример же — что-нибудь да значит. Я хочу быть там, где опасно, хочу пойти навстречу опасности, — и безоружным.
Мне сказали, что необходим пропуск, подписанный вами. Прошу вас послать мне его.
Будьте уверены, генерал, в моих самых лучших к вам чувствах.
В. Гюго.
`
const msgAboutID = `
Происхождение выражения «ехать зайцем» не удаётся отследить. Вполне возможно, что оно перешло в русский из французского (en lapin, дословно — кроликом), из которого уже, кажется, исчезло! Самое раннее упоминание пока что нашлось у Чехова В вагоне 1881-го:

«На железных дорогах зайцами называются гг. пассажиры, затрудняющие разменом денег не кассиров, а кондукторов. Хорошо, читатель, ездить зайцем! Зайцам полагается, по нигде еще не напечатанному тарифу, 75% уступки, им не нужно толпиться около кассы, вынимать ежеминутно из кармана билет, с ними кондуктора вежливее и... всё что хотите, одним словом!»

А у Бальзака в Первых шагах в жизни 1844-го, призванных увековечить экипажи-кукушки, «процветавшие в течение целого столетия, многочисленные еще и в 1830 году, которые уже не встретишь, разве только в день сельского праздника»:

«На передке имелись широкие деревянные козлы, предназначенные для Пьеротена; рядом с ним могло усесться еще три пассажира: таких пассажиров, как известно, называют “зайцами” (qui prennent, comme on le sait, le nom de lapins). Случалось, что Пьеротен брал на козлы и четырех “зайцев”, а сам тогда примащивался сбоку на чем-то вроде ящика, приделанного снизу к кузову и наполненного соломой или такими посылками, в которые “зайцы” могли без опаски упираться ногами.»

– Évidemment, dit le clerc, le comte est le voyageur qui sans l’obligeance d’un jeune homme allait se mettre en lapin dans la voiture à Pierrotin.
– En lapin, dans la voiture à Pierrotin ?... s’écrièrent le régisseur et la fille de basse-cour.

«Уже недалеко то время, когда их вытеснят железные дороги». В 1872-м еженедельник Универсальный музей [подтверждает это](https://gallica.bnf.fr/ark:/12148/bpt6k5775615p/f389.item.r=lapins.zoom), юмористически и без толики сожаления — их кукушки были как наши маршрутки.

Во французском слово заяц ассоциировалось больше с местом, чем со стоимостью проезда, и не прикрепилось к пассажирам поездов… Однако ещё оставалось жить! В 1923-м Маленький парижанин в заметке Le coucou et les lapins (под [центральной картинкой](https://gallica.bnf.fr/ark:/12148/bpt6k605596h/f2.highres), [текстом](https://gallica.bnf.fr/ark:/12148/bpt6k605596h.textePage.f2.langFR)) задаётся вопросом, почему пассажиры такси говорят своим спутникам, что поедут зайцами, когда уступают им удобные уголки. (Кажется, неудобным при этом считается место между другими, а не рядом с водителем или где-то ещё.) Для ответа находится общий момент — и новым, и старым зайцам не комфортно сидеть рядом с кем-то (с кучером) — а название заяц объясняется позой, которую занимает такой пассажир. Если это верно, то понятно, почему lapin в этом смысле больше не встречается.
`

var tmplFuncs = template.FuncMap{"num": num, "plu": plu, "joinWithAnd": joinWithAnd}

var tmpl = template.Must(template.New("").Funcs(tmplFuncs).Parse(`
{{ define "hello" -}}
Привет! Я помощник для игры в [Зайца](https://boardgamegeek.com/boardgame/227072/chameleon)!
{{- end -}}

{{ define "Words" -}}
	{{- if .Private -}}
		Добавь меня в группу и напиши {{""}}
	{{- else -}}
		Напишите {{""}}
        {{- end -}}
` + "`/hare слова`" + `, чтобы начать. (В оригинале 16 слов на одну тему; [примеры](http://ref.manuscriptgames.org/chameleon/index.html).) Отдельные слова можно вводить через пробел, а словосочетания через запятую или с новой строки.
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
{{- template "add" . }}
/rules — правила
/about — обо мне
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
