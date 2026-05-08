.PHONY: test e2e completions

test:
	go test ./...

e2e:
	docker compose -f test/e2e/docker-compose.yml up --abort-on-container-exit --exit-code-from e2e

completions:
	mkdir -p completions
	go run . completion bash > completions/outline.bash
	go run . completion zsh > completions/outline.zsh
	go run . completion fish > completions/outline.fish
