.PHONY: swagger-client dev check-deps

check-deps:
	@command -v yq >/dev/null 2>&1 || { echo >&2 "yq is required but it's not installed. Aborting."; exit 1; }

PRINTER_ENDPOINT := $(shell yq eval '.printer.endpoint' config.toml)

dev: check-deps swagger-client
	go run ./cmd/go-thermal-printer-discord/main.go

swagger-client: check-deps
	rm -rf ./internal/swagger
	mkdir -p ./internal/swagger
	swagger generate client -f "$(PRINTER_ENDPOINT)/swagger/doc.json" -A printer -t ./internal/swagger
