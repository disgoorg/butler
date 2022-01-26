FROM golang:1.17.6-alpine AS build

WORKDIR /tmp/app

COPY . .

RUN apk add --no-cache git && \
    go mod download && \
    go mod verify && \
    go build -o butler

FROM alpine:latest

WORKDIR /home/disgo-butler

COPY --from=build /tmp/app/disgo-butler /home/disgo-butler/

EXPOSE 80

ENTRYPOINT ./disgo-butler