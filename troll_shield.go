// Copyright 2020 the commonlispbr authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var trollGroups = []string{"@ccppbrasil", "@progclucb", "@progclucb"}

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
		if chatMember.IsMember() {
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

func main() {
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
						"Kicking %v did not work, error code %v: %v",
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

					bot.Send(msg)
				}
			}
		}
	}
}
