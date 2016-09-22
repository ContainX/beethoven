GO_FMT = gofmt -s -w -l .

all: deps compile

compile:
	go build ./...

deps:
	go get

docker-dev:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=0.1-DEV -X main.built=$(date -u '+%Y-%m-%d %H:%M:%S')"
	docker build -t containx/beethoven .

format:
	$(GO_FMT)
