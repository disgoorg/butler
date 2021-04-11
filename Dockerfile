FROM golang:1.16.2-alpine AS build

WORKDIR /tmp/app

COPY . .

RUN apk add --no-cache git && \
    go mod download && \
    go mod verify && \
    go build -o disgo-butler

FROM alpine:latest

WORKDIR /home/disgo-butler

COPY --from=build /tmp/app/disgo-butler /home/disgo-butler/

EXPOSE 80

ENTRYPOINT ./disgo-butler