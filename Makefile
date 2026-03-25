BINARY := project-pipe
CONFIG  := config.yaml

.PHONY: run build tidy

run:
	CONFIG_PATH=$(CONFIG) go run ./cmd/server

build:
	go build -o $(BINARY) ./cmd/server

tidy:
	go mod tidy
