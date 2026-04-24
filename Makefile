.PHONY: build clean test vet lint typeset-deps

# Binary name = repo name = service name on the VM.
BINARY := jdbbs

build:
	go build -o $(BINARY) ./cmd/srv

clean:
	rm -f $(BINARY)

test:
	go test ./...

vet:
	go vet ./...

# Install Python deps used by typesetting/scripts/*.py
typeset-deps:
	pip3 install --user python-docx pyyaml
