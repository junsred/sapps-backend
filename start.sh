#!/bin/bash
export GOTRACEBACK=all
export GODEBUG=asyncpreemptoff=1

# Import .env file
if [ -f /root/sapps/.env ]; then
    export $(cat /root/sapps/.env | grep -v '^#' | xargs)
fi

cd /root/sapps/services/go
go build -o /tmp/sapps-app cmd/sapps/main.go
exec /tmp/sapps-app 