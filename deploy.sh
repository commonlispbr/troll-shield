#!/bin/bash
set -eux
server=lisp@bikelis
go build
ssh $server 'pkill troll-shield || true'
scp troll-shield $server:
git describe --tags | ssh $server 'cat > version.txt'
ssh bikelis 'tmux new-session -d -s troll || true'
ssh bikelis 'tmux new-window -d ./run.sh'
