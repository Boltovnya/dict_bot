package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	rg "github.com/go-redis/redis/v8"
	tb "gopkg.in/telebot.v3"
)

var ctx = context.Background()

func main() {

	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	rdb := rg.NewClient(&rg.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/define", func(h tb.Context) error {
		m := h.Message()
		var r, q, hd string
		splitString := strings.Fields(m.Payload)

		if len(splitString) < 1 {
			r = "Are you gonna tell me what you want?"
		} else {
			q = strings.Join(splitString[:1], " ")
			hd = strings.ToLower(q)
			val, err := rdb.Get(ctx, hd).Result()

			if err == rg.Nil {
				r = fmt.Sprintf("I can't find \"%s\" for some reason 🤔", q)
				log.Printf("%s - %s", err, m.Payload)
			} else if err != nil {
				r = fmt.Sprintf("I can't find \"%s\" for some reason 🤔", q)
				log.Printf("%s - %s", err, m.Payload)
			} else {
				r = fmt.Sprintf("%s - %s", q, val)
			}
		}
		return h.Reply(r, &tb.ReplyMarkup{})
		
	})

	b.Handle("/definenew", func(h tb.Context) error {
		m := h.Message()
		var r string
		splitString := strings.Fields(m.Payload)

		if len(splitString) <= 1 {
			r = "I need something to define it as... 🙄"
		} else {
			key := strings.ToLower(strings.Join(splitString[:1], " "))
			definition := strings.Join(splitString[1:], " ")
			_, exists := rdb.Get(ctx, key).Result()

			if exists == rg.Nil {
				err := rdb.Set(ctx, key, definition, 0).Err()
				if err != nil {
					r = "Something went wrong when creating definition, Please try again later"
				} else {
					r = fmt.Sprintf("Definition added! Thanks @%s++", m.Sender.Username)
				}
			} else {
				r = "Definition already exists! If you wish to replace it, delete it with `/rmdef $Definition` first."
			}
		}
		return h.Reply(r, &tb.ReplyMarkup{})
	})

	b.Handle("/rmdef", func(h tb.Context) error {
		m := h.Message()
		var r string
		splitString := strings.Fields(m.Payload)
		
		if len(splitString) < 1 {
			r = "I need something to delete... 🙄"
		}
		key := strings.ToLower(strings.Join(splitString[:1], " "))
		_, exists := rdb.Get(ctx, key).Result()
		if exists == rg.Nil {
			r = "Nothing to delete here 🎉"
		} else {
			rdb.Del(ctx, key)
			r = "Zap ⚡️, it's gone"
		}
		return h.Reply(r, &tb.ReplyMarkup{})
	})

	b.Start()
}
