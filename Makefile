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

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...

docker-build:
	docker build -t $(APPLICATION_NAME) .

docker-run:
	docker run -it --rm -p 9290:9290 $(APPLICATION_NAME)

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/$(APPLICATION_NAME)-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/$(APPLICATION_NAME)-arm64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/$(APPLICATION_NAME)-freebsd-386 main.go
