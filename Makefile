build:
	@go build ./...
.PHONY: build

tidy:
	@go mod tidy
.PHONY: tidy

generate:
	@buf generate
.PHONY: generate
