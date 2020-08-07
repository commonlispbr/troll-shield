import json
import os
import telepot  # notype
from tqdm import tqdm  # notype

token = open("token.txt").read().strip()
bot = telepot.Bot(token)
chat_title = "Common Lisp Brasil"
logs_fpath = "logs/putaria.log.json"
dir_name = os.path.join("docs", chat_title.replace(" ", "_").lower())


def get_title(result):
    return result.get("message", {}).get("chat", {}).get("title")


def collect_documents(chat_title=chat_title):
    doc_types = [
        "document",
        "video",
        "voice",
        "photo"
    ]
    mime_type = {
        "video": "video/mp4",
        "voice": "audio/ogg",
        "photo": "image/jpg",
    }
    logs = json.load(open(logs_fpath))
    docs = []
    for timestamp, event in logs.items():
        for result in event["result"]:
            if get_title(result) == chat_title:
                for doc_type in doc_types:
                    doc = result["message"].get(doc_type)
                    if isinstance(doc, list):
                        doc = doc[-1]  # multiple thumbs, get the best quality
                        doc["mime_type"] = mime_type[doc_type]
                    if doc:
                        docs.append(doc)

    return docs


def collect_messages(chat_title=chat_title):
    logs = json.load(open(logs_fpath))
    docs = []
    for timestamp, event in logs.items():
        for result in event["result"]:
            if get_title(result) == chat_title:
                doc = result["message"]
                if doc:
                    docs.append({
                        "date": timestamp,
                        "message": doc
                    })
    return docs


def dump_messages(messages):
    with open(os.path.join(dir_name, "messages.txt"), "w") as f:
        for message in sorted(messages, key = lambda x: x["date"]):
            date = message["date"]
            msg = message["message"]
            username = message["message"]["from"]["first_name"]
            text = message["message"].get("text")
            if text:
                template = f"{date} / {username}: {text}".replace("\n", " ")
                f.write(template + "\n")


def download_document(doc, dir_name=dir_name):
    try:
        mime_type = doc.get("mime_type")
        extension = ".raw"
        if mime_type:
            extension = mime_type.replace("/", ".")
        elif doc.get("file_name"):
            extension = doc["file_name"]
        fname = doc["file_unique_id"] + extension
        folder = mime_type.split("/")[0]
        dir_path = os.path.join(dir_name, folder)
        os.makedirs(dir_path, exist_ok=True)
        fpath = os.path.join(dir_name, folder, fname)
        if not os.path.exists(fpath):
            bot.download_file(doc["file_id"], fpath)
    except telepot.exception.TelegramError as e:
        print(f"Telegram exception for {fname}: {e}")
    except Exception as e:
        print(f"Python exception, I screw up: {e}")


def download_documents(docs):
    for doc in tqdm(docs):
        download_document(doc, dir_name)


if __name__ == "__main__":
    os.makedirs(dir_name, exist_ok=True)
    print(f"-- Collecting documents in: {dir_name}")
    docs = collect_documents()
    download_documents(docs)
    print(f"-- Collecting messages in: {dir_name}")
    messages = collect_messages()
    dump_messages(messages)
