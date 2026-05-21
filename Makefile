BINARY    := outline
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null)
DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GOFLAGS   := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build test vet e2e completions docs dev-up dev-down dev-logs clean help

all: vet test build

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -o $(BINARY) .

test:
	go test -v -count=1 ./...

vet:
	go vet ./...

e2e:
	docker compose -f test/e2e/docker-compose.yml up --abort-on-container-exit --exit-code-from e2e

completions:
	mkdir -p completions
	go run . completion bash > completions/outline.bash
	go run . completion zsh > completions/outline.zsh
	go run . completion fish > completions/outline.fish

docs:
	go run . docs generate
	@echo "Generated: docs/reference/generated/, llms.txt, llms-full.txt, .github/copilot-instructions.md, CLAUDE.md, .cursor/rules"

dev-up:
	docker compose -f dev/docker-compose.yml up -d

dev-down:
	docker compose -f dev/docker-compose.yml down

dev-logs:
	docker compose -f dev/docker-compose.yml logs -f

clean:
	rm -f $(BINARY)

help:
	@echo "Targets:"
	@echo "  all          - vet, test, build (default)"
	@echo "  build        - build binary"
	@echo "  test         - run unit tests"
	@echo "  vet          - run go vet"
	@echo "  e2e          - run e2e tests"
	@echo "  completions  - generate shell completions"
	@echo "  docs         - generate all docs (reference, llms.txt, LLM instructions)"
	@echo "  dev-up       - start dev environment"
	@echo "  dev-down     - stop dev environment"
	@echo "  dev-logs     - tail dev logs"
	@echo "  clean        - remove binary"
	@echo "  help         - show this help"
