// Copyright 2020 the commonlispbr authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	logger "log"
	"os"
	"strings"
	"sync"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TrollShieldBot aggregate the methods used by my bot to keep mocking easier
type TrollShieldBot interface {
	GetChatMember(telegram.ChatConfigWithUser) (telegram.ChatMember, error)
	KickChatMember(telegram.KickChatMemberConfig) (telegram.APIResponse, error)
	Send(telegram.Chattable) (telegram.Message, error)
}

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

// messageEvent return true if is a message event
func messageEvent(update *telegram.Update) bool {
	return update.Message != nil
}

// newChatMemberEvent return true if a new member joined to chat
func newChatMemberEvent(update *telegram.Update) bool {
	return messageEvent(update) && update.Message.NewChatMembers != nil
}

// fromChatEvent return true if the message is from a specific chat
func fromChatEvent(update *telegram.Update, username string) bool {
	chat := update.Message.Chat
	return messageEvent(update) && chat != nil && (chat.UserName == username || chat.Title == username)
}

// getUserName return the most meaningful name available from a telegram user
// if it has a @username, return it
// if not, try to return FirstName + LastName
// otherwise, only return the FirstName
func getUserName(user telegram.User) string {
	username := user.FirstName
	if user.UserName != "" {
		username = fmt.Sprintf("@%v", user.UserName)
	} else if user.LastName != "" {
		username = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return username
}

func getUpdates(bot *telegram.BotAPI) telegram.UpdatesChannel {
	u := telegram.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("getUpdates error: %v", err)
	}

	return updates
}

func reply(bot TrollShieldBot, update *telegram.Update, text string) {
	msg := telegram.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("[!] Send msg failed: %v", err)
	}
}

func welcomeMessage(bot TrollShieldBot, update *telegram.Update, member telegram.User) {
	username := getUserName(member)
	text := fmt.Sprintf(
		`Olá %s! Seja bem-vindo ao grupo oficial de Common Lisp do Brasil.
Leia as regras em: https://lisp.com.br/rules.html.`,
		username,
	)
	reply(bot, update, text)
}

// findTrollHouses return a string with groups separeted by comma,
// that groups are well-known to being troll houses.
// otherwise, if nothing is found returns a empty string
func findTrollHouses(bot TrollShieldBot, userID int) string {
	ch := make(chan string, len(trollGroups))
	var wait sync.WaitGroup
	for _, trollGroup := range trollGroups {
		wait.Add(1)
		go func(group string) {
			defer wait.Done()
			c, _ := bot.GetChatMember(telegram.ChatConfigWithUser{
				SuperGroupUsername: group,
				UserID:             userID,
			})
			if c.IsMember() || c.IsCreator() || c.IsAdministrator() {
				ch <- group
			} else {
				ch <- ""
			}
		}(trollGroup)

	}
	go func() {
		wait.Wait()
		close(ch)
	}()
	var houses []string
	for house := range ch {
		if house != "" {
			houses = append(houses, house)
		}
	}

	return strings.Join(houses, ", ")
}

// kickTroll ban the troll and send a message about where we can found the trolls
func kickTroll(bot TrollShieldBot, update *telegram.Update, user telegram.User, trollHouse string) {
	chatMember := telegram.ChatMemberConfig{
		ChatID: update.Message.Chat.ID,
		UserID: user.ID,
	}
	resp, err := bot.KickChatMember(
		telegram.KickChatMemberConfig{ChatMemberConfig: chatMember},
	)

	if resp.Ok == false || err != nil {
		log.Printf(
			"[!] Kicking %q did not work, error code %v: %v",
			user.FirstName, resp.ErrorCode, resp.Description,
		)
	} else {
		username := getUserName(user)
		text := fmt.Sprintf(
			"%v foi banido porque é membro do grupo: %v. Adeus.",
			username, trollHouse,
		)
		reply(bot, update, text)
	}
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
	err = telegram.SetLogger(log)
	if err != nil {
		log.Printf("Set Telegram Bot Logging error: %v", err)
	}
}

func setupBot() *telegram.BotAPI {
	token, exists := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !exists {
		log.Fatal("TELEGRAM_BOT_TOKEN env should be defined.")
	}
	bot, err := telegram.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account @%s", bot.Self.UserName)

	return bot
}

func main() {
	setupLogging()
	bot := setupBot()
	for update := range getUpdates(bot) {
		if messageEvent(&update) {
			if update.Message.Text == "/lelerax" {
				reply(bot, &update, "Estou vivo.")
			}
		}

		if newChatMemberEvent(&update) {
			for _, member := range *update.Message.NewChatMembers {
				if trollHouse := findTrollHouses(bot, member.ID); trollHouse != "" {
					kickTroll(bot, &update, member, trollHouse)
				} else if fromChatEvent(&update, "commonlispbr") && !member.IsBot {
					welcomeMessage(bot, &update, member)
				}
			}
		}
	}
}
