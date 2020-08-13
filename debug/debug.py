# coding: utf-8

from telepot import Bot
from pprint import pprint

bot = Bot(open("token.txt").read().strip())

def get_chat_debug():
    users = {
        "lerax": 336558555,
        "tretanews": 430571154,
    }

    chats = {
        "commonlispbr": -1001280636766,
        "commonlisphq": -1001493125566,
    }

    chat_member = bot.getChatMember(chats["commonlisphq"],
                                    users["tretanews"])
    print()
    pprint(chat_member)


def set_bot_commands():
    commands = [
        {
            "command": "kills",
            "description": "Retorna a quantidade de trolls decapitados."
        },
        {
            "command": "lelerax",
            "description": "Ping. Verifica se estou vivo."
        },
        {
            "command": "pass",
            "description": "Commando de passe /pass <@username>. Só funciona com os admins do @commonlispbr."
        }
    ]

    bot.setMyCommands(commands)
    pprint(bot.getMyCommands())
