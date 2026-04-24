.PHONY: build clean stop start restart test

build:
	go build -o prodcal ./cmd/srv

clean:
	rm -f prodcal

test:
	go test ./...
