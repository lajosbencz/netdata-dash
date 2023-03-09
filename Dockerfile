FROM golang:1.19-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o demo-server .

FROM debian:10-slim

RUN apt update && apt install wget -y && \
    wget -O /tmp/netdata-kickstart.sh https://my-netdata.io/kickstart.sh && \
    sh /tmp/netdata-kickstart.sh --no-updates --stable-channel --disable-telemetry \
    --disable-cloud --native-only

COPY --from=builder /app/demo-server /usr/bin/
COPY ./run.sh /app/run.sh
RUN chmod +x /app/run.sh
EXPOSE 9301
EXPOSE 9302
EXPOSE 19999
VOLUME [ "/etc/demo-server/" ]
ENTRYPOINT [ "/app/run.sh" ]
