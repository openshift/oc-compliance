BINDIR=./bin
BIN=$(BINDIR)/oc-compliance

export GOFLAGS=-mod=vendor

SRC=$(shell find . -name *.go)

.PHONY: all
all: build

.PHONY: build
build: $(BIN)

$(BIN): $(BINDIR) $(SRC)
	go build -o $(BIN) github.com/JAORMX/oc-compliance/cmd

.PHONY: install
install: build
	which oc | xargs dirname | xargs -n1 cp $(BIN)

# Helper targets

$(BINDIR):
	mkdir -p $(BINDIR)
