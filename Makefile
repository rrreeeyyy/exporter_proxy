COMMIT = $(shell git describe --always)
VERSION = $(shell grep Version cli/version.go | sed -E 's/.*"(.+)"$$/\1/')
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GOXOSARCH ?= "darwin/amd64 linux/amd64 linux/arm64 windows/amd64"

default: build

# build generate binary on './bin' directory.
build:
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/exporter_proxy .

buildx:
	gox -ldflags "-X main.GitCommit=$(COMMIT)" -output "bin/v$(VERSION)/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch $(GOXOSARCH) .

lint:
	golint ${GOFILES_NOVENDOR}

vet:
	go vet -v ${GOFILES_NOVENDOR}

test:
	go test

test-short:
	go test -short

fmt:
	gofmt -l -w ${GOFILES_NOVENDOR}

release: buildx
	git tag v$(VERSION)
	git push origin v$(VERSION)
	ghr v$(VERSION) bin/v$(VERSION)/

