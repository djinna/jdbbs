.PHONY: build clean stop start restart test local

# Version shown on the factory step strip (0.33): MMDD.shorthash, e.g. 0922.9258e56
VERSION := $(shell git log -1 --format=%cd --date=format:%m%d 2>/dev/null || echo 0000).$(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

build:
	go build -ldflags "-X srv.exe.dev/srv.BuildVersion=$(VERSION)" -o prodcal ./cmd/srv

clean:
	rm -f prodcal

stop:
	sudo systemctl stop prodcal

start:
	sudo systemctl start prodcal

restart:
	sudo systemctl restart prodcal

test:
	go test ./...

local:
	./scripts/run-local.sh
