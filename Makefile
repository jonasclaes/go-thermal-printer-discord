.PHONY: swagger-client dev

dev: swagger-client
	go run ./cmd/go-thermal-printer-discord/main.go

swagger-client:
	rm -rf ./internal/swagger
	mkdir -p ./internal/swagger
	swagger generate client -f "${PRINTER_BASE_URL}/swagger/doc.json" -A printer -t ./internal/swagger
