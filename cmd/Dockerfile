FROM golang:1.18-alpine AS build

ARG VERSION=dev

WORKDIR /tmp/app

COPY . .

RUN go mod download && \
    go mod verify && \
    go build -ldflags="-X 'main.version=${VERSION}'" -o disgo-butler cmd/main.go

FROM alpine

WORKDIR /app

COPY --from=build /tmp/app/disgo-butler /app/disgo-butler

ENTRYPOINT ["./disgo-butler"]
