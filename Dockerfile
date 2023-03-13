FROM golang:1.19-alpine as builder
ARG cmd=netdata-dash
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
ENV GODEBUG=http2client=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
RUN go test -v ./...
RUN go build -ldflags '-w -s' -o ${cmd} /app/cmd/${cmd}

FROM scratch
ARG cmd=netdata-dash
COPY --from=builder /app/${cmd} /usr/bin/
EXPOSE 16666
WORKDIR /etc/netdata-dash/
VOLUME [ "/etc/netdata-dash/" ]
ENTRYPOINT [ "${cmd}" ]
