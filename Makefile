PROJECT_ROOT=$(shell pwd)

test:
	@go test ./...

lint:
	@docker run --rm -v $(PROJECT_ROOT):/app -w /app golangci/golangci-lint:v1.42.0 golangci-lint run -v