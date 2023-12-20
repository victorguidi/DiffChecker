run: build
	@./docfiff -h

build:
	@go build -v -o docfiff ./src/

.PHONY: all run build
