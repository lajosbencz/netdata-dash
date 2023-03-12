.PHONY: dev dev-agent

all: dash dash-agent

dash:
	go build ./make/netdata-dash/

dash-agent:
	go build ./make/netdata-dash-agent/

dash-client:
	go build ./make/netdata-dash-client/

dev:
	go run ./make/netdata-dash/

dev-agent:
	go run ./make/netdata-dash-agent/

dev-client:
	go run ./make/netdata-dash-client/
