.PHONY: build install clean

build:
	go build -o ssm ./cmd/ssm

install: build
	cp ssm ~/.local/bin/ssm

clean:
	rm -f ssm
