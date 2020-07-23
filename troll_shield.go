// Copyright 2020 the commonlispbr authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	logger "log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// blacklist groups, member from that groups will be kicked automatically
var trollGroups = []string{
	"@ccppbrasil",
	"@vaicaraiooo",
	"@javascriptbr",
	"@frontendbr",
	"@WebDevBR",
	"@progclucb",
	"@progclube",
	"@commonlispbrofficial",
}

const logfile = "troll-shield.log"

var log = logger.New(os.Stderr, "", logger.LstdFlags)

// findTrollHouse return the troll house group name if is well-known
// otherwise, returns a empty string
func findTrollHouse(bot *tgbotapi.BotAPI, userID int) (string, error) {
	var error error
	for _, trollGroup := range trollGroups {
		c, err := bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: trollGroup,
			UserID:             userID,
		})
		if err != nil {
			error = err
			continue
		}
		if c.IsMember() || c.IsCreator() || c.IsAdministrator() {
			return trollGroup, nil
		}
	}

	return "", error
}

// messageEvent: return true if is a message event
func messageEvent(update *tgbotapi.Update) bool {
	return update.Message != nil
}

// newChatMemberEvent: return true if a new member joined to chat
func newChatMemberEvent(update *tgbotapi.Update) bool {
	return messageEvent(update) && update.Message.NewChatMembers != nil
}

// fromChatEvent: return true if the message is from a specific chat
func fromChatEvent(update *tgbotapi.Update, username string) bool {
	chat := update.Message.Chat
	return messageEvent(update) && chat != nil && (chat.UserName == username || chat.Title == username)
}

func setupLogging() {
	// log to console and file
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)

	log.SetOutput(wrt)
	// register log to BotLoggt
	err = tgbotapi.SetLogger(log)
	if err != nil {
		log.Printf("Set Telegram Bot Logging error: %v", err)
	}
}

func getUserName(user tgbotapi.User) string {
	username := user.FirstName
	if user.UserName != "" {
		username = fmt.Sprintf("@%v", user.UserName)
	} else if user.LastName != "" {
		username = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return username
}

func reply(bot *tgbotapi.BotAPI, update *tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("[!] Send msg failed: %v", err)
	}
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

	log.Printf("Authorized on account @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if messageEvent(&update) {
			if update.Message.Text == "/lelerax" {
				reply(bot, &update, "Estou vivo.")
			}
		}

		if newChatMemberEvent(&update) {
			for _, member := range *update.Message.NewChatMembers {
				trollHouse, err := findTrollHouse(bot, member.ID)
				if err != nil {
					log.Printf("[!] findTrollHouse returned a error: %v", err)
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
							"[!] Kicking %q did not work, error code %v: %v",
							member.FirstName, resp.ErrorCode, resp.Description,
						)
					} else {
						username := getUserName(member)
						text := fmt.Sprintf(
							"%v foi banido porque é membro do grupo: %v. Adeus.",
							username, trollHouse,
						)
						reply(bot, &update, text)
					}
				} else {
					if fromChatEvent(&update, "commonlispbr") && !member.IsBot {
						username := getUserName(member)
						text := fmt.Sprintf(
							`Olá %s! Seja bem-vindo ao grupo oficial de Common Lisp do Brasil.
Leia as regras em: https://lisp.com.br/rules.html.`,
							username,
						)
						reply(bot, &update, text)
					}
				}
			}
		}
	}
}
