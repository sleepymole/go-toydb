build:
	@go build ./...
.PHONY: build

tidy:
	@go mod tidy
.PHONY: tidy

generate:
	@cd api && buf generate
.PHONY: generate

fmt:
	@go fmt ./...
	@cd api && buf format -w
.PHONY: fmt
