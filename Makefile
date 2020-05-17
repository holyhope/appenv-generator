BINARY ?= ./bin/generator

install:
	$(MAKE) build BINARY="$${HOME}/bin/generator"

build: $(BINARY)

tests: $(BINARY)
	# https://github.com/ovh/venom/
	venom run ./tests/venom.yaml --var binary=$(BINARY)

$(BINARY): ./cmd/*.go ./markers/*.go ./v1/*.go ./generator/*.go
	go build -o $(BINARY) cmd/*.go
