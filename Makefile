
GOBASE= $(shell pwd)
GOSRC= $(GOBASE)/src
GOBIN= $(GOBASE)/bin
GOFILES= $(wildcard *.go)

.PHONY: build  
build:
	go build -C $(GOSRC) -o $(GOBIN)/ 

.PHONY: linter
linter:
	cd src/; golangci-lint run
