.PHONY: build clean stop start restart test local

build:
	go build -o prodcal ./cmd/srv

clean:
	rm -f prodcal

test:
	go test ./...

local:
	./scripts/run-local.sh
