all: netdata-dash netdata-dash-agent

netdata-dash-agent:
	go build ./netdata-dash-agent/

netdata-dash:
	go build .

dev:
	go run .

dev-agent:
	go run ./netdata-dash-agent/
