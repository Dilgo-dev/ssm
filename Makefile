.PHONY: build install clean fmt lint check

build:
	go build -o ssm ./cmd/ssm

install: build
	cp ssm ~/.local/bin/ssm

clean:
	rm -f ssm

fmt:
	gofmt -w .

lint:
	golangci-lint run ./...

check: fmt lint build
