// Copyright 2020 the commonlispbr authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	logger "log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TrollShieldBot aggregate the methods used by my bot to keep mocking easier
type TrollShieldBot interface {
	GetChatMember(telegram.ChatConfigWithUser) (telegram.ChatMember, error)
	KickChatMember(telegram.KickChatMemberConfig) (telegram.APIResponse, error)
	UnbanChatMember(telegram.ChatMemberConfig) (telegram.APIResponse, error)
	Send(telegram.Chattable) (telegram.Message, error)
	LeaveChat(telegram.ChatConfig) (telegram.APIResponse, error)
	GetUpdatesChan(telegram.UpdateConfig) (telegram.UpdatesChannel, error)
}

// blacklist groups, member from that groups will be kicked automatically
var trollGroups = []string{
	"@ccppbrasil",
	"@vaicaraiooo",
	"@progclube",
	"@commonlispbrofficial",
	"@mlbrasil",
}

var passList = []string{}

var admins = []string{
	"lerax",
	"luksamuk",
	"perkunos",
	"renan_r",
}

const logfile = "troll-shield.log"
const killsFile = "kills.txt"

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

func getUpdates(bot TrollShieldBot) telegram.UpdatesChannel {
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
func kickTroll(bot TrollShieldBot, update *telegram.Update, user telegram.User, trollHouse string) error {
	chatMember := telegram.ChatMemberConfig{
		ChatID: update.Message.Chat.ID,
		UserID: user.ID,
	}
	resp, err := bot.KickChatMember(
		telegram.KickChatMemberConfig{
			ChatMemberConfig: chatMember,
			UntilDate:        time.Now().AddDate(0, 0, 1).Unix(), // one day
			//UntilDate: time.Now().Add(time.Minute * 1).Unix(), // one minute
		},
	)

	if !resp.Ok || err != nil {
		log.Printf(
			"[!] Kicking %q did not work, error code %v: %v",
			user.FirstName, resp.ErrorCode, resp.Description,
		)
	} else {
		username := getUserName(user)
		text := fmt.Sprintf(
			"%v foi removido porque é membro do grupo: %v. Para mais informações, acione o nosso SAC 24h: @skhaz.",
			username, trollHouse,
		)
		reply(bot, update, text)
	}

	return err
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

func setupBot(envVar string) (*telegram.BotAPI, error) {
	token, exists := os.LookupEnv(envVar)
	if !exists {
		return nil, fmt.Errorf("%s env should be defined", envVar)
	}
	bot, err := telegram.NewBotAPI(token)

	if err != nil {
		return nil, fmt.Errorf("setup %v failed with: %v", envVar, err)
	}

	bot.Debug = true

	log.Printf("Authorized on account @%s", bot.Self.UserName)

	return bot, nil
}

func setupHiddenBot(bot *telegram.BotAPI) *telegram.BotAPI {
	log.Println("Setup the hidden bot")
	botHidden, err := setupBot("TELEGRAM_BOT_HIDDEN_TOKEN")
	if err != nil {
		log.Printf("Bot setup failed: %v. Fallback to main bot.", err)
		botHidden = bot
	}

	return botHidden

}

func setupBots() (*telegram.BotAPI, *telegram.BotAPI, error) {
	log.Println("Setup the main bot")
	bot, err := setupBot("TELEGRAM_BOT_TOKEN")
	if err != nil {
		return nil, nil, err
	}

	return bot, setupHiddenBot(bot), nil
}

func leaveChat(bot TrollShieldBot, update *telegram.Update, trollGroup string) {
	reply(bot, update, "Nesse grupo há trolls. Dou-me a liberdade de ir embora. Adeus.")
	r, err := bot.LeaveChat(telegram.ChatConfig{ChatID: update.Message.Chat.ID})
	if !r.Ok || err != nil {
		log.Printf("Bot tried to exit from %v, but failed with: %v",
			trollGroup, err,
		)
	}
}

func loadKills(fpath string) int64 {
	dat, err := ioutil.ReadFile(fpath)
	if err == nil {
		i, err := strconv.Atoi(strings.TrimSpace(string(dat)))
		if err != nil {
			log.Printf("Parsing %q go bad, got error: %v", fpath, err)
		} else {
			return int64(i)
		}
	}

	return 0
}

func saveKills(fpath string, kills int64) error {
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		_, err = f.WriteString(strconv.FormatInt(kills, 10))
	}
	if e := f.Close(); e != nil {
		err = e
	}
	return err
}

func reportKills(bot TrollShieldBot, update *telegram.Update, kills int64) {
	txt := fmt.Sprintf("%v foram sacrificados.", kills)
	if kills%2 == 0 {
		txt = fmt.Sprintf("Já taquei o pau em %v trolls!", kills)
	}
	reply(bot, update, txt)
}

// parse:
// - /pass <@username>
// - /pass <FirstName [LastName]>
// and return what is available
func extractPassUserName(command string) string {
	tokens := strings.Split(command, " ")
	n := len(tokens)
	return strings.Join(tokens[1:n], " ")
}

// If has pass, return true and and return the matched pass
func hasPass(user telegram.User) (string, bool) {
	userName := getUserName(user)
	for _, pass := range passList {
		if strings.HasPrefix(userName, pass) || user.FirstName == pass {
			return pass, true
		}
	}
	return "", false
}

// remove pass list and reply the consumed pass list
func removePassList(bot TrollShieldBot, update *telegram.Update, pass string) {
	// Remove the element at index i from a.
	n := len(passList)
	for i, p := range passList {
		if p == pass {
			// remove pass from passList
			passList[i] = passList[n-1] // Copy last element to index i.
			passList[n-1] = ""          // Erase last element (write zero value).
			passList = passList[:n-1]   // Truncate slice.
		}
	}
	reply(bot, update, fmt.Sprintf("O passe para %q foi consumido.", pass))
}

// check if a message cames from a @commonlispbr admin
func fromAdminEvent(update *telegram.Update) bool {
	if update.Message.From == nil {
		return false
	}
	fromUserName := update.Message.From.UserName
	for _, admin := range admins {
		if admin == fromUserName {
			return true
		}
	}

	return false
}

// addPass to passList and send a message
func addPassList(bot TrollShieldBot, update *telegram.Update) {
	userName := extractPassUserName(update.Message.Text)
	if len(userName) > 0 {
		passList = append(passList, userName)
		reply(bot, update, fmt.Sprintf("O passe para %q foi adicionado.", userName))
	}
}

func commandEvent(update *telegram.Update) bool {
	return messageEvent(update) && strings.HasPrefix(update.Message.Text, "/")
}

// check if a /command is valid
func checkCommand(botUserName string, msg string, command string) bool {
	isQualified := strings.Contains(msg, command+"@")
	if isQualified {
		qualifiedCommand := fmt.Sprintf("%v@%v", command, botUserName)
		return strings.HasPrefix(msg, qualifiedCommand)
	}
	return strings.HasPrefix(msg, command)
}
