ifeq ($(shell git tag --contains HEAD),)
  VERSION := $(shell git rev-parse --short HEAD)
else
  VERSION := $(shell git tag --contains HEAD)
endif

BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS += -X github.com/dennigogo/miniweb/general.Version=$(VERSION)
GOLDFLAGS += -X github.com/dennigogo/miniweb/general.BuildTime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

.PHONY: build

build: ## Build application
	GOSUMDB=off go build -o miniweb $(GOFLAGS) -v ./main.go

docker-start: ## Run containers
	docker build -t miniweb --compress -f Dockerfile ./
	docker-compose up -d miniweb