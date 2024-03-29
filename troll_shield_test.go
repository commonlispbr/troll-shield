package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BotMockup struct{}

func (bot *BotMockup) GetChatMember(c telegram.ChatConfigWithUser) (telegram.ChatMember, error) {
	switch c.UserID {
	case 1:
		return telegram.ChatMember{Status: "member"}, nil
	case 2:
		return telegram.ChatMember{Status: "creator"}, nil
	case 3:
		return telegram.ChatMember{Status: "administrator"}, nil
	case 4:
		return telegram.ChatMember{Status: "left"}, nil
	default:
		return telegram.ChatMember{}, errors.New("user not found")
	}
}

func (bot *BotMockup) KickChatMember(c telegram.KickChatMemberConfig) (telegram.APIResponse, error) {
	switch c.ChatMemberConfig.UserID {
	case 0:
		return telegram.APIResponse{Ok: true}, nil
	default:
		return telegram.APIResponse{Ok: false}, errors.New("error")
	}

}

func (bot *BotMockup) UnbanChatMember(c telegram.ChatMemberConfig) (telegram.APIResponse, error) {
	switch c.UserID {
	case 0:
		return telegram.APIResponse{Ok: true}, nil
	default:
		return telegram.APIResponse{Ok: false}, errors.New("error")
	}

}

func (bot *BotMockup) Send(c telegram.Chattable) (telegram.Message, error) {
	return telegram.Message{}, nil
}

func (bot *BotMockup) LeaveChat(c telegram.ChatConfig) (telegram.APIResponse, error) {
	switch c.ChatID {
	case 1:
		return telegram.APIResponse{Ok: true}, nil
	default:
		return telegram.APIResponse{Ok: false}, errors.New("user not found")
	}
}

func (bot *BotMockup) GetUpdatesChan(c telegram.UpdateConfig) (telegram.UpdatesChannel, error) {
	return make(chan telegram.Update, 1), nil
}

func TestGetUserName(t *testing.T) {
	user1 := telegram.User{
		FirstName: "Rolisvaldo",
	}
	if got := getUserName(user1); got != "Rolisvaldo" {
		t.Errorf("getUserName when only FirstName is available should return it, not %v", got)
	}

	user2 := telegram.User{
		FirstName: "Rolisvaldo",
		LastName:  "Da Silva",
	}
	if got := getUserName(user2); got != "Rolisvaldo Da Silva" {
		t.Errorf("getUserName when FirstName and LastName are available should return it, not %v", got)
	}

	user3 := telegram.User{
		FirstName: "Rolisvaldo",
		LastName:  "Da Silva",
		UserName:  "rolisvaldo",
	}
	if got := getUserName(user3); got != "@rolisvaldo" {
		t.Errorf("getUserName when only UserName is available should return it with @, not %v", got)
	}
}

func TestEvents(t *testing.T) {
	// messageEvent
	update := telegram.Update{}
	if got := messageEvent(&update); got != false {
		t.Errorf("messageEvent should return false when there is no message, got: %v", got)
	}
	message := telegram.Message{}
	update.Message = &message
	if got := messageEvent(&update); got != true {
		t.Errorf("messageEvent should return true when there is a message, got: %v", got)
	}

	// newChatMemberEvent
	if got := newChatMemberEvent(&update); got != false {
		t.Errorf("newChatMemberEvent should return false when there is no new members, got: %v", got)
	}
	newChatMembers := []telegram.User{}
	update.Message.NewChatMembers = &newChatMembers
	if got := newChatMemberEvent(&update); got != true {
		t.Errorf("newChatMemberEvent should return true when there is new members, got: %v", got)
	}

	// fromChatEvent
	if got := fromChatEvent(&update, "commonlispbr"); got != false {
		t.Errorf("fromChatEvent should return false when there is no new members, got: %v", got)
	}
	chat := telegram.Chat{UserName: "commonlispbr", Title: "CommonLispBR HQ"}
	update.Message.Chat = &chat
	if got := fromChatEvent(&update, "commonlispbr"); got != true {
		t.Errorf("fromChatEvent should return true when there is a chat with UserName, got: %v", got)
	}
	if got := fromChatEvent(&update, "CommonLispBR HQ"); got != true {
		t.Errorf("fromChatEvent should return true when there is new members, got: %v", got)
	}

	// commandEvent
	update.Message.Text = "pancadaria tiro e bomba"
	if got := commandEvent(&update); got != false {
		t.Errorf("commandEvent should return false when there is no prefix / on Message.Text, got: %v", got)
	}
	update.Message.Text = "/command"
	if got := commandEvent(&update); got != true {
		t.Errorf("commandEvent should return true when receives /command like message, got: %v", got)
	}

}

func TestFindTrollHouses(t *testing.T) {
	botnilson := BotMockup{}
	trollGroups = []string{"@rolisvaldo"}
	if got := findTrollHouses(&botnilson, 1); got != "@rolisvaldo" {
		t.Errorf("findTrollHouses expects @rolisvaldo, got: %v", got)
	}
	if got := findTrollHouses(&botnilson, 2); got != "@rolisvaldo" {
		t.Errorf("findTrollHouses expects @rolisvaldo, got: %v", got)
	}
	if got := findTrollHouses(&botnilson, 3); got != "@rolisvaldo" {
		t.Errorf("findTrollHouses expects @rolisvaldo, got: %v", got)
	}
	if got := findTrollHouses(&botnilson, 4); got != "" {
		t.Errorf("findTrollHouses expects empty string, got: %v", got)
	}
	if got := findTrollHouses(&botnilson, -1); got != "" {
		t.Errorf("findTrollHouses expects empty string, got: %v", got)
	}
}

func TestKickTroll(t *testing.T) {
	botnilson := BotMockup{}
	update := telegram.Update{}
	message := telegram.Message{}
	chat := telegram.Chat{}
	message.Chat = &chat
	update.Message = &message
	user := telegram.User{}
	if err := kickTroll(&botnilson, &update, user, "@trollhouse"); err != nil {
		t.Errorf("kickTroll error: %v", err)
	}
	user.ID = 1
	if err := kickTroll(&botnilson, &update, user, "@trollhouse"); err == nil {
		t.Errorf("kickTroll should fail, but got: %v", err)
	}
}

func TestWelcomeMessage(t *testing.T) {
	botnilson := BotMockup{}
	update := telegram.Update{}
	message := telegram.Message{}
	chat := telegram.Chat{}
	message.Chat = &chat
	update.Message = &message
	user := telegram.User{}
	welcomeMessage(&botnilson, &update, user)
}

func TestSetupBot(t *testing.T) {
	envVar := "TELEGRAM_BOT_TOKEN"
	if err := os.Setenv(envVar, "123"); err != nil {
		t.Errorf("Setup env var TELEGRAM_BOT_TOKEN error: %v", err)
	}

	if _, err := setupBot(envVar); err == nil {
		t.Errorf("Invalid token should go fail, got nil error")
	}

	if _, err := setupBot("???"); err == nil {
		t.Errorf("Non-defined env var should go fail, got nil error.")
	}
}

func TestSetupBots(t *testing.T) {
	bot, _, err := setupBots()
	if err == nil {
		t.Errorf("setupBots fail with invalid tokens.")
	}
	botHidden := setupHiddenBot(bot)
	if botHidden != bot {
		t.Errorf("When botHidden fails to start, use bot as fallback")
	}
}

func TestSetupLogging(t *testing.T) {
	setupLogging()
}

func TestLeaveChat(t *testing.T) {
	bot := BotMockup{}
	update := telegram.Update{}
	message := telegram.Message{}
	chat := telegram.Chat{}
	message.Chat = &chat
	update.Message = &message
	leaveChat(&bot, &update, "trolleira")
}

func TestGetUpdates(t *testing.T) {
	bot := BotMockup{}
	getUpdates(&bot)
}

func TestSaveLoadKills(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "kills.txt")
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("isso-nao-eh-um-numero")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if kills := loadKills(tmpfile.Name()); kills != 0 {
		t.Errorf("loadKills should return 0 when there is a invalid number, got: %v", kills)
	}

	// write 10
	if err := saveKills(tmpfile.Name(), 10); err != nil {
		t.Errorf("saveKills failed: %v", err)
	}

	// read 10
	if kills := loadKills(tmpfile.Name()); kills != 10 {
		t.Errorf("Save/Load of KillCounter didn't worked, expected 10, got %v", kills)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(tmpfile.Name()); err != nil {
		t.Fatal(err)
	}

}

func TestReportKills(t *testing.T) {
	bot := BotMockup{}
	update := telegram.Update{}
	message := telegram.Message{}
	chat := telegram.Chat{}
	message.Chat = &chat
	update.Message = &message
	reportKills(&bot, &update, int64(11))
	reportKills(&bot, &update, int64(10))
}

func TestExtractPassUserName(t *testing.T) {
	tableTest := []struct {
		input    string
		expected string
	}{
		{
			"/pass @lerax",
			"@lerax",
		},
		{
			"/pass First Name",
			"First Name",
		},
		{
			"/pass@lelerax_bot @tretanews_bot",
			"@tretanews_bot",
		},
		{
			"/pass",
			"",
		},
	}

	for _, test := range tableTest {
		if got := extractPassUserName(test.input); got != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, got)
		}
	}
}

func TestPassList(t *testing.T) {

	bot := BotMockup{}
	update := telegram.Update{}
	message := telegram.Message{}
	chat := telegram.Chat{}
	message.Chat = &chat
	message.Text = "/pass @lerax"
	update.Message = &message
	user := telegram.User{UserName: "lerax"}

	// adding test
	addPassList(&bot, &update)
	t.Logf("passList: %v", passList)
	if pass, ok := hasPass(user); pass != "@lerax" && ok != true {
		t.Errorf("User @lerax should have a pass: pass=%v, ok=%v", pass, ok)
	}

	// removing test
	t.Logf("passList: %v", passList)
	removePassList(&bot, &update, "@lerax")
	if pass, ok := hasPass(user); ok != false {
		t.Errorf("User @lerax should not have more a pass: pass=%v, ok=%v", pass, ok)
	}
}

func TestFromAdminEvent(t *testing.T) {
	update := telegram.Update{}
	message := telegram.Message{}
	user := telegram.User{UserName: "lerax"}
	message.From = &user
	update.Message = &message
	if got := fromAdminEvent(&update); got == false {
		t.Errorf("lerax is an eternal admin, it should be true")
	}

	user.UserName = "delduca"
	if got := fromAdminEvent(&update); got == true {
		t.Errorf("delduca should not even being a member, neither admin.")
	}
}

func TestCheckCommand(t *testing.T) {
	tableTest := []struct {
		botUserName string
		msg         string
		command     string
		expected    bool
	}{
		{
			"lelerax_bot",
			"/kills@lelerax_bot",
			"/kills",
			true,
		},
		{
			"botnilson",
			"/pass @infeliz",
			"/pass",
			true,
		},
		{
			"botsvaldo",
			"/pass@lelerax_bot @username",
			"/pass",
			false,
		},
		{
			"botsvaldo",
			"/pass@botsvaldo @username",
			"/pass",
			true,
		},
	}

	for i, test := range tableTest {
		if got := checkCommand(test.botUserName, test.msg, test.command); got != test.expected {
			t.Errorf("%d. Expected %v, got %v for %+v.", i+1, test.expected, got, test)
		}
	}

}
