.PHONY: build doc fmt lint dev test vet godep install bench

PKG_NAME=$(shell basename `pwd`)

install:
	go get -t -v ./...

build: vet \
	test \
	go build -v -o ./bin/$(PKG_NAME)

doc:
	godoc -http=:6060

fmt:
	go fmt ./...

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint ./...

dev:
	DEBUG=* go get && go install && gin -p 8911 -i

test:
	go test ./...

# Runs benchmarks
bench:
	go test ./... -bench=.

# https://godoc.org/golang.org/x/tools/cmd/vet
vet:
	go vet ./...

godep:
	godep save ./...
