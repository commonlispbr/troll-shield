from telepot import Bot
from pprint import pprint

bot = Bot(open("token.txt").read().strip())
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
