import re
import enum
import sys
import json

log_end_regex = re.compile(r"^(\d{4}/\d{2}/\d{2}.*).*")
log_start_regex = re.compile(r"^(\d{4}/\d{2}/\d{2}.*).getUpdates.resp: ({.*)")

class ParsingState(enum.Enum):
    stop = 0
    start = 1


def parse(fpath: str):
    logs = {}
    state_machine = ParsingState.stop
    with open(fpath) as f:
        key = ""
        for line in f.readlines():
            if log_start_regex.match(line):
                match = log_start_regex.search(line)
                key = match.group(1)
                value = match.group(2)
                logs[key] = value
                state_machine = ParsingState.start
            elif log_end_regex.match(line):
                state_machine = ParsingState.stop
            elif state_machine == ParsingState.start:
                logs[key] += line

    for k, v in sorted(logs.items(), key = lambda x: x[1]):
        try:
            j = json.loads(v)
            if not j or not j["result"]:
                del logs[k]
            logs[k] = j
        except Exception as e:
            del logs[k]
    return logs


if __name__ == '__main__':
    fname = sys.argv[1]
    logs = parse(fname)
    json.dump(logs, open(fname + ".json", "w"))
