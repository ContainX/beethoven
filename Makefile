GO_FMT = gofmt -s -w -l .

all: deps compile

compile:
	go build ./...

deps:
	go get

docker-dev:
	GOOS=linux GOARCH=amd64 go build
	docker build -t beethoven .

format:
	$(GO_FMT)
