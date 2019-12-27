FROM golang:1.13.4-alpine3.10

RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
    git \
    go get github.com/gorilla/websocket
