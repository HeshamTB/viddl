
all: build

build:
	go build .

watch-deps:
	go install github.com/cosmtrek/air@latest

watch: watch-deps
	air


