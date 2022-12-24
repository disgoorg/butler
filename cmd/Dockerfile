FROM golang:1.19-alpine AS build

ARG VERSION=dev

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o bot cmd/main.go

FROM alpine

COPY --from=build /build/bot /bin/bot

ENTRYPOINT ["/bin/bot"]
