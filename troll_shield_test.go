package main

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TestGetUserName(t *testing.T) {
	user1 := tgbotapi.User{
		FirstName: "Rolisvaldo",
	}
	if got := getUserName(user1); got != "Rolisvaldo" {
		t.Errorf("getUserName when only FirstName is available should return it, not %v", got)
	}

	user2 := tgbotapi.User{
		FirstName: "Rolisvaldo",
		LastName:  "Da Silva",
	}
	if got := getUserName(user2); got != "Rolisvaldo Da Silva" {
		t.Errorf("getUserName when FirstName and LastName are available should return it, not %v", got)
	}

	user3 := tgbotapi.User{
		FirstName: "Rolisvaldo",
		LastName:  "Da Silva",
		UserName:  "rolisvaldo",
	}
	if got := getUserName(user3); got != "@rolisvaldo" {
		t.Errorf("getUserName when only UserName is available should return it with @, not %v", got)
	}
}
