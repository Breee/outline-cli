.PHONY: test e2e

test:
	go test ./...

e2e:
	docker compose -f test/e2e/docker-compose.yml up --abort-on-container-exit --exit-code-from e2e
