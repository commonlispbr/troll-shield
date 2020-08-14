//+build !test
// Copyright 2020 the commonlispbr authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
)

func main() {
	setupLogging()
	bot, botHidden, err := setupBots()
	botUser := bot.Self.UserName
	if err != nil {
		log.Fatal(err.Error())
	}
	kills := loadKills(killsFile)
	log.Printf("Currently kill state: %v", kills)
	for update := range getUpdates(bot) {
		if messageEvent(&update) {
			// Exit automatically from group after the bot receive a message from it
			for _, trollGroup := range trollGroups {
				if fromChatEvent(&update, strings.TrimLeft(trollGroup, "@")) {
					leaveChat(bot, &update, trollGroup)
				}
			}
		}

		if commandEvent(&update) {
			msg := update.Message.Text
			if checkCommand(botUser, msg, "/ping") {
				reply(bot, &update, "Estou vivo.")
			}

			if checkCommand(botUser, msg, "/kills") {
				reportKills(bot, &update, kills)
			}

			if checkCommand(botUser, msg, "/pass") && fromAdminEvent(&update) {
				addPassList(bot, &update)
			}
		}

		if newChatMemberEvent(&update) {
			for _, member := range *update.Message.NewChatMembers {
				if pass, ok := hasPass(member); ok {
					removePassList(bot, &update, pass)
					welcomeMessage(bot, &update, member)
					continue
				}

				if trollHouse := findTrollHouses(botHidden, member.ID); trollHouse != "" {
					err := kickTroll(bot, &update, member, trollHouse)
					if err == nil {
						kills++
						if err := saveKills(killsFile, kills); err != nil {
							log.Printf("saving kills failed: %v", err)
						}
					}
				} else if fromChatEvent(&update, "commonlispbr") && !member.IsBot {
					welcomeMessage(bot, &update, member)
				}

				// Exit automatically from groups when I'm joining it
				for _, trollGroup := range trollGroups {
					if fromChatEvent(&update, strings.TrimLeft(trollGroup, "@")) && member.UserName == bot.Self.UserName {
						leaveChat(bot, &update, trollGroup)
					}
				}

			}
		}
	}
}
