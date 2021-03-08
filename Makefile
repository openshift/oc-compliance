BINDIR=./bin
BIN=$(BINDIR)/oc-compliance

export GOFLAGS=-mod=vendor

SRC=$(shell find . -name *.go)

OSCAP=$(shell which oscap || echo /usr/bin/oscap)

.PHONY: all
all: build

.PHONY: build
build: $(BIN)

$(BIN): $(BINDIR) $(SRC)
	go build -o $(BIN) github.com/JAORMX/oc-compliance/cmd

.PHONY: install
install: build
	which oc | xargs dirname | xargs -n1 cp $(BIN)


.PHONY: e2e
e2e: install dependencies
	go test ./tests/e2e -timeout 40m -v --ginkgo.v


# Helper targets

$(BINDIR):
	mkdir -p $(BINDIR)

.PHONY: dependencies
dependencies: $(OSCAP)

$(OSCAP):
	sudo yum install -y openscap
