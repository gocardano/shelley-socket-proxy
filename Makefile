PROJECT_NAME=shelley-socket-proxy
PROJECT_SRCDIR=github.com/gocardano/${PROJECT_NAME}

VERSION="0.0.1-$(shell git rev-parse --short=8 HEAD)"
DOCKER_ARGS="--rm -u $(shell id -u) -e GOCACHE=/tmp/"

GOLANG_IMAGE="golang:1.15.2"
GOLINT_IMAGE="golangci/golangci-lint:v1.31.0"

.PHONY: default fmt vet test coverage build

default: container

fmt:
	@echo ➭ Running go fmt
	@docker run "${DOCKER_ARGS}" -v `pwd`:/go/src/${PROJECT_SRCDIR} \
		-w /go/src/${PROJECT_SRCDIR} ${GOLANG_IMAGE} \
		go fmt ./... | read 1>&2 && exit 1 || true

vet:
	@echo ➭ Running go vet
	@docker run "${DOCKER_ARGS}" -v `pwd`:/go/src/${PROJECT_SRCDIR} \
		-w /go/src/${PROJECT_SRCDIR} ${GOLANG_IMAGE} go vet ./...

lint:
	@echo ➭ Running go lint
	@docker run --rm -v `pwd`:/app \
		-w /app ${GOLINT_IMAGE} golangci-lint run -v

test:
	@echo ➭ Running go test
	@docker run "${DOCKER_ARGS}" \
		-v `pwd`:/go/src/${PROJECT_SRCDIR} \
		-w /go/src/${PROJECT_SRCDIR} ${GOLANG_IMAGE} go test ./...

coverage:
	@echo ➭ Running go test coverage
	@docker run "${DOCKER_ARGS}" \
		-v `pwd`:/go/src/${PROJECT_SRCDIR} \
		-w /go/src/${PROJECT_SRCDIR} ${GOLANG_IMAGE} go test -coverprofile=.cov ./...;  go tool cover -func .cov

coverage-html: coverage
	@go tool cover -html=.cov

build: fmt vet lint
	@echo ➭ Building ${PROJECT_NAME}
	@docker run "${DOCKER_ARGS}" \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e GO111MODULE=off \
		-e CGO_ENABLED=0 \
		-v `pwd`:/go/src/${PROJECT_SRCDIR} \
		-w /go/src/${PROJECT_SRCDIR} ${GOLANG_IMAGE} go build \
			-o ${PROJECT_NAME} \
			-ldflags "-X main.version=${VERSION}" \
			github.com/gocardano/${PROJECT_NAME}/cmd/cli
