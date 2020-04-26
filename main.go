package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/orivej/e"
	tb "gopkg.in/tucnak/telebot.v2"
)

type NoPoller struct{}

func (NoPoller) Poll(*tb.Bot, chan tb.Update, chan struct{}) {}

const envToken = "OddHareGameBotToken"

var flPoll = flag.Bool("poll", false, "poll for updates")
var flDebug = flag.Bool("debug", false, "enable debug logging")
var flToken = flag.String("token", os.Getenv(envToken), "bot token: id:key in $"+envToken)
var flName = flag.String("name", "OddHareGameBot", "bot name")

func main() {
	flag.Parse()
	if *flToken == "" {
		fmt.Println("error: missing -token or $" + envToken)
		os.Exit(1)
	}
	seed()

	p := tb.Poller(NoPoller{})
	if *flPoll {
		p = &tb.LongPoller{Timeout: 10 * time.Second}
	}
	if *flDebug {
		p = tb.NewMiddlewarePoller(p, debugFilter)
	}

	b, err := tb.NewBot(tb.Settings{Token: *flToken, Poller: p})
	e.Exit(err)
	NewBot(b).Setup()
	// if !*flPoll {
	// 	lambda.Start(func(u tb.Update) error {
	// 		b.Updates <- u
	// 		return nil
	// 	})
	// }
	b.Start()
}

func debugFilter(u *tb.Update) bool {
	s, err := json.Marshal(u)
	e.Print(err)
	log.Printf("%s", s)
	return true
}
