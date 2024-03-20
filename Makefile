build:
	@go build ./...
.PHONY: build

tidy:
	@go mod tidy
.PHONY: tidy

generate:
	@buf generate
.PHONY: generate

fmt:
	@go fmt ./...
	@buf format -w
.PHONY: fmt
