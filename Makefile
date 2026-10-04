.PHONY: run local test build docker

run:    ## SSH on :23234
	go run ./cmd/ssh-portfolio

local:  ## TUI in this terminal, no SSH
	go run ./cmd/ssh-portfolio -local

test:
	go vet ./... && go test ./...

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ssh-portfolio ./cmd/ssh-portfolio

docker:
	docker build -t ssh-portfolio .
