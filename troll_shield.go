// Copyright 2020 the commonlispbr authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// blacklist groups, member from that groups will be kicked automatically
var trollGroups = []string{
	"@ccppbrasil",
	"@vaicaraiooo",
	"@javascriptbr",
	"@frontendbr",
	"@GuiaDev",
	"@WebDevBR",
	"@mundojs",
	"@progclucb",
	"@progclube",
	"@commonlispbrofficial",
}

const logfile = "troll-shield.log"

// findTrollHouse return the troll house group name if is well-known
// otherwise, returns a empty string
func findTrollHouse(bot *tgbotapi.BotAPI, userID int) (string, error) {
	var error error = nil
	for _, trollGroup := range trollGroups {
		chatMemberConf := tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: trollGroup,
			UserID:             userID,
		}
		chatMember, err := bot.GetChatMember(chatMemberConf)
		if err != nil {
			error = err
			continue
		}
		if chatMember.IsMember() || chatMember.IsCreator() || chatMember.IsAdministrator() {
			return trollGroup, nil
		}
	}

	return "", error
}

// TODO: should only works on @commonlispbr on future
// selectedEvent return true if is a desired event to be processed
func selectedEvent(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.NewChatMembers != nil
}

func setupLogging() {
	// log to console and file
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)

	log.SetOutput(wrt)
}

func main() {
	setupLogging()
	token, exists := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !exists {
		log.Fatal("TELEGRAM_BOT_TOKEN env should be defined.")
	}
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		// Check if is join event
		if !selectedEvent(&update) { // ignore any non-Message Updates
			continue
		}

		for _, member := range *update.Message.NewChatMembers {
			trollHouse, err := findTrollHouse(bot, member.ID)
			if err != nil {
				log.Printf("findTrollHouse betrayed us: %v", err)
				continue
			}

			if trollHouse != "" {
				chatMember := tgbotapi.ChatMemberConfig{
					ChatID: update.Message.Chat.ID,
					UserID: member.ID,
				}
				resp, err := bot.KickChatMember(
					tgbotapi.KickChatMemberConfig{ChatMemberConfig: chatMember},
				)

				if resp.Ok == false || err != nil {
					log.Printf(
						"Kicking %q did not work, error code %v: %v",
						member.FirstName, resp.ErrorCode, resp.Description,
					)
				} else {
					username := member.FirstName
					if member.UserName != "" {
						username = fmt.Sprintf("@%v", member.UserName)
					}
					text := fmt.Sprintf(
						"%v foi banido porque é membro do grupo: %v. Adeus.",
						username, trollHouse,
					)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					msg.ReplyToMessageID = update.Message.MessageID

					_, err := bot.Send(msg)
					if err != nil {
						log.Printf("Send msg failed: %v", err)
					}
				}
			}
		}
	}
}
