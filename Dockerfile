FROM golang AS builder

# pre-copy/cache go.mod for pre-downloading dependencies 
# and only redownloading them in subsequent builds if they change
WORKDIR /usr/src/bot
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# too much random bullsh#t in root directory
# dont wanna even bother with dockerignore
COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/

# need to mount output directory...
# FROM builder as unittest
# RUN go test -coverprofile=coverage.out ./...
# RUN go tool cover -html=coverage.out -o cover.html

FROM builder as botbuild
RUN CGO_ENABLED=0 go build -o ./cmd/thebot/bot ./cmd/thebot/main.go

FROM builder as mockbuild
RUN CGO_ENABLED=0 go build -o ./cmd/tgmock/mock ./cmd/tgmock/main.go

FROM alpine AS bot
COPY --from=botbuild /usr/src/bot/cmd/thebot/bot .
COPY ./configs/config.bot.yaml ./config.yaml
ENTRYPOINT [ "./bot", "--config", "config.yaml" ]

FROM alpine AS mock
COPY --from=mockbuild /usr/src/bot/cmd/tgmock/mock .
COPY ./configs/config.mock.yaml ./config.yaml
ENTRYPOINT [ "./mock", "--config", "config.yaml" ]
