FROM python:3.8.0-alpine3.10

RUN apk update \
  && apk add --no-cache \
  git \
  && pip3 install mahjong \
  && pip3 install git+https://github.com/Pithikos/python-websocket-server \
  && pip3 install websocket-server
