package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/orivej/OddHareGameBot/bot"
	"github.com/orivej/e"
	tb "gopkg.in/tucnak/telebot.v2"
)

const envDebug = "OddHareGameBotDebug"
const envToken = "OddHareGameBotToken"

var flPoll = flag.Bool("poll", false, "poll for updates")
var flDebug = flag.Bool("debug", os.Getenv(envDebug) != "", "enable debug logging")
var flToken = flag.String("token", os.Getenv(envToken), "bot token: id:key in $"+envToken)
var flName = flag.String("name", "OddHareGameBot", "bot name")
var flLocal = flag.Bool("local", false, "keep state in local memory")
var flTable = flag.String("ddbtable", "OddHareGameBotTable", "DynamoDB table name")

func main() {
	flag.Parse()
	if *flToken == "" {
		fmt.Println("error: missing -token or $" + envToken)
		os.Exit(1)
	}

	cfg := tb.Settings{Token: *flToken}
	if *flPoll {
		cfg.Poller = tb.Poller(&tb.LongPoller{Timeout: 10 * time.Second})
		if *flDebug {
			cfg.Poller = tb.NewMiddlewarePoller(cfg.Poller, debugFilter)
		}
	} else {
		cfg.Updates = 1
		cfg.Synchronous = true
	}

	b, err := tb.NewBot(cfg)
	e.Exit(err)
	bot.NewBot(b, *flName, *flLocal, *flTable).Setup()
	if *flPoll {
		b.Start()
	} else {
		type Request struct {
			Body string `json:"body"`
		}
		lambda.Start(func(req Request) error {
			var u tb.Update
			err = json.Unmarshal([]byte(req.Body), &u)
			e.Exit(err)
			if *flDebug {
				debugFilter(&u)
			}
			b.ProcessUpdate(u)
			return nil
		})
	}
}

func debugFilter(u *tb.Update) bool {
	s, err := json.Marshal(u)
	e.Print(err)
	log.Printf("%s", s)
	return true
}
