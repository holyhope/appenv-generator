BINARY ?= ./bin/generator

SOURCE_FILES := $(wildcard ./cmd/*.go) $(wildcard ./markers/*.go) $(wildcard ./v1/*.go) $(wildcard ./generator/*.go)
TEST_FILES := $(filter-out ./tests/zz_generated.appenv.go, $(wildcard ./tests/*.go))

install:
	$(MAKE) build BINARY="$${HOME}/bin/generator"

build: $(BINARY)

tests: $(BINARY) ./venom.yaml $(TEST_FILES)
	# https://github.com/ovh/venom/
	venom run ./venom.yaml --var binary=$(BINARY)

$(BINARY): $(SOURCE_FILES)
	go build -o $(BINARY) ./cmd/main.go
