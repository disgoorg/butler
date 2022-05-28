FROM golang:1.18-alpine AS build

ARG VERSION=dev

WORKDIR /tmp/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-X 'main.version=${VERSION}'-w -s" -o disgo-butler cmd/main.go

FROM alpine

COPY --from=build /tmp/app/disgo-butler /bin/disgo-butler

ENTRYPOINT ["/bin/disgo-butler"]
