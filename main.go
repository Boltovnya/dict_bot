package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	rg "github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

var ctx = context.Background()

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("error loading .env file")
	}

	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	rdb := rg.NewClient(&rg.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/define", func(m *tb.Message) {
		splitString := strings.Fields(m.Payload)
		if len(splitString) < 1 {
			b.Reply(m, "Err... what do you want?", &tb.ReplyMarkup{})
			return
		}
		query := strings.Join(splitString[:1], " ")
		head := strings.ToLower(query)
		val, err := rdb.Get(ctx, head).Result()
		if err == rg.Nil {
			b.Reply(m, fmt.Sprintf("I can't find \"%s\" for some reason 🤔", query), &tb.ReplyMarkup{})
			log.Printf("%s - %s", err, m.Payload)
		} else if err != nil {
			b.Reply(m, fmt.Sprintf("Something went wrong when finding \"%s\"", query), &tb.ReplyMarkup{})
		} else {
			b.Reply(m, fmt.Sprintf("%s - %s", query, val), &tb.ReplyMarkup{})
		}
	})

	b.Handle("/definenew", func(m *tb.Message) {
		splitString := strings.Fields(m.Payload)
		if len(splitString) <= 1 {
			b.Reply(m, "I need something to define it as... 🙄", &tb.ReplyMarkup{})
			return
		}
		key := strings.ToLower(strings.Join(splitString[:1], " "))
		definition := strings.Join(splitString[1:], " ")

		_, exists := rdb.Get(ctx, key).Result()
		if exists == rg.Nil {
			err := rdb.Set(ctx, key, definition, 0).Err()
			if err != nil {
				b.Reply(m, "Something went wrong when creating definition. Please try again later", &tb.ReplyMarkup{})
			}
			b.Reply(m, fmt.Sprintf("Definition added! Thanks @%s++", m.Sender.Username), &tb.ReplyMarkup{})
		} else {
			b.Reply(m, "Definition already exists! If you wish to replace it, delete it with `/rmdef $Definition`", &tb.ReplyMarkup{})
		}
	})

	b.Handle("/rmdef", func(m *tb.Message) {
		splitString := strings.Fields(m.Payload)
		if len(splitString) < 1 {
			b.Reply(m, "I need something to define it as... 🙄", &tb.ReplyMarkup{})
			return
		}
		key := strings.ToLower(strings.Join(splitString[:1], " "))
		_, exists := rdb.Get(ctx, key).Result()
		if exists == rg.Nil {
			b.Reply(m, "Nothing to delete here 🎉", &tb.ReplyMarkup{})
			return
		}
		rdb.Del(ctx, key)
		b.Reply(m, "Zap ⚡️, it's gone", &tb.ReplyMarkup{})
	})

	b.Start()
}
