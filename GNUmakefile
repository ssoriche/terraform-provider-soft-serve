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

.PHONY: default build install test testacc lint generate
