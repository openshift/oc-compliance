BINDIR=./bin
BIN=$(BINDIR)/oc-fcr

export GOFLAGS=-mod=vendor

SRC=$(wildcard cmd/*go)

.PHONY: all
all: build

.PHONY: build
build: $(BIN)

$(BIN): $(BINDIR) $(SRC)
	go build -o $(BIN) github.com/JAORMX/fetch-compliance-results/cmd

.PHONY: install
install: build
	which oc | xargs dirname | xargs -n1 cp $(BIN)

# Helper targets

$(BINDIR):
	mkdir -p $(BINDIR)
