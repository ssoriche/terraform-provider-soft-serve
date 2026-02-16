default: build

build:
	go build -o terraform-provider-soft-serve

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/ssoriche/soft-serve/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-soft-serve ~/.terraform.d/plugins/registry.terraform.io/ssoriche/soft-serve/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

lint:
	golangci-lint run ./...

generate:
	go generate ./...

docs:
	tfplugindocs generate --provider-name soft-serve

fmt:
	go fmt ./...

clean:
	rm -f terraform-provider-soft-serve
	rm -rf dist/

help:
	@echo "Available targets:"
	@echo "  build     - Build the provider binary"
	@echo "  install   - Build and install to local Terraform plugins directory"
	@echo "  test      - Run unit tests"
	@echo "  testacc   - Run acceptance tests (requires running Soft Serve instance)"
	@echo "  lint      - Run golangci-lint"
	@echo "  generate  - Run go generate"
	@echo "  docs      - Generate provider documentation"
	@echo "  fmt       - Format Go source code"
	@echo "  clean     - Remove build artifacts"
	@echo "  help      - Show this help message"

.PHONY: default build install test testacc lint generate docs fmt clean help
