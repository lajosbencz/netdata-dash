FROM golang:1.19-alpine as builder
ARG cmd=netdata-dash
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GODEBUG=http2client=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o ${cmd} /app/cmd/${cmd}

FROM scratch
ARG cmd=netdata-dash
COPY --from=builder /app/${cmd} /usr/bin/
EXPOSE 16666
WORKDIR /etc/netdata-dash/
VOLUME [ "/etc/netdata-dash/" ]
ENTRYPOINT [ "${cmd}" ]
