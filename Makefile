

HOSTNAME=terraform-example.com
NAMESPACE=haproxy-provider
NAME=haproxy
APPLICATION_NAME=terraform-provider-${NAME}
VERSION=1.0.0
GOARCH=darwin_arm64
INSTALL_PATH=~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOARCH}
INSTALL_PATH=/Users/cepitacio/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOARCH}/

default: build

build:
	go build -o $(APPLICATION_NAME) .
	mv $(APPLICATION_NAME) $(INSTALL_PATH)

build_local:
	go build -o $(APPLICATION_NAME) .
	mv $(APPLICATION_NAME) $(INSTALL_PATH)
	rm -rf examples/resources/.terraform* && rm -rf examples/resources/terraform*
	cd examples/resources && terraform init

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

# Testing targets
test:
	go test -count=1 -parallel=4 ./...

test-verbose:
	go test -count=1 -parallel=4 -v ./...

test-race:
	go test -count=1 -race -parallel=4 ./...

test-coverage:
	go test -count=1 -parallel=4 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-benchmark:
	go test -count=1 -bench=. -benchmem ./...

# Acceptance tests (require real HAProxy instance)
testacc:
	@echo "Running acceptance tests..."
	@echo "Make sure you have HAProxy running with Data Plane API enabled"
	@echo "Set environment variables: HAPROXY_ENDPOINT, HAPROXY_API_VERSION"
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 30m -v ./...





# Linting and code quality
lint:
	@echo "Running linters..."
	golangci-lint run

lint-fix:
	@echo "Running linters with auto-fix..."
	golangci-lint run --fix

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

# Security scanning
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping security scan"; \
		echo "Run 'make install-tools-alt' to install gosec"; \
	fi

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/hashicorp/terraform-plugin-testing@latest

# Alternative installation methods (no GitHub auth required)
install-tools-alt:
	@echo "Installing development tools (alternative method)..."
	@echo "Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.64.8
	@echo "Installing gosec..."
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

	@echo "Installing terraform-plugin-testing..."
	go install github.com/hashicorp/terraform-plugin-testing@latest

# Clean up
clean:
	@echo "Cleaning up..."
	rm -f terraform-provider-haproxy
	rm -f coverage.out coverage.html
	rm -rf bin/
	rm -rf .terraform/
	rm -rf examples/*/.terraform/
	rm -rf examples/*/terraform.tfstate*

# Full test suite
test-all: fmt vet lint test test-coverage
	@echo "All tests completed successfully!"

# CI/CD pipeline
ci: install-tools fmt vet lint test test-coverage security
	@echo "CI pipeline completed successfully!"

docker-build:
	docker build -t $(APPLICATION_NAME) .

docker-run:
	docker run -it --rm -p 9290:9290 $(APPLICATION_NAME)

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/$(APPLICATION_NAME)-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/$(APPLICATION_NAME)-arm64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/$(APPLICATION_NAME)-freebsd-386 main.go
