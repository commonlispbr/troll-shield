#!/bin/bash
set -eux
server=lisp@bikelis
CGO_ENABLED=0 go build  -a -ldflags '-extldflags "-static"' .
ssh $server 'pkill troll-shield || true'
scp troll-shield $server:
git describe --tags | ssh $server 'cat > version.txt'
ssh bikelis 'tmux new-session -d -s troll || true'
ssh bikelis 'tmux new-window -d ./run.sh'
