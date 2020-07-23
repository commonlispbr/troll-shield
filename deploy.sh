#!/bin/bash
set -eux
go build
ssh bikelis 'pkill troll-shield || true'
scp troll-shield bikelis:
git describe --tags | ssh bikelis 'cat > version.txt'
ssh bikelis 'tmux new-session -d -s troll'
ssh bikelis 'tmux new-window -d ./run.sh'
