.EXPORT_ALL_VARIABLES:

ifndef VERSION
VERSION := $(shell git describe --always --tags)
endif

DATE := $(shell date -u +%Y%m%d.%H%M%S)

LDFLAGS = -trimpath -ldflags "-X=main.version=$(VERSION)-$(DATE)"
CGO_ENABLED=0

targets = mbtilestokv kvtilesd 

.PHONY: all lint test clean mbtilestokv testnolint kvtilesd

all: test $(targets)

test: lint testnolint

CGO_ENABLED=1
testnolint:
	go test -race ./...

lint:
	golangci-lint run

kvtilesd:
	cd cmd/kvtilesd && go build $(LDFLAGS)

mbtilestokv: CGO_ENABLED=1
mbtilestokv:
	cd cmd/mbtilestokv && go build $(LDFLAGS)

clean:
	rm -f cmd/mbtilestokv/mbtilestokv
	rm -f cmd/kvtilesd/kvtilesd
