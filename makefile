modules=services/go
include .env
export $(shell sed 's/=.*//' .env)
app:
	echo "Running Humanize"
	cd services/go && /bin/sh -c "go build -o /tmp/sapps-app cmd/sapps/main.go && exec /tmp/sapps-app"
notifications:
	echo "Running Notifications"
	cd services/go && /bin/sh -c "go build -o /tmp/notifications-app cmd/notifications/main.go && exec /tmp/notifications-app"
tidy:
	for module in $(modules) ; do \
        cd $$module && go mod tidy; \
        go mod tidy; \
        cd ../../ ; \
    done