CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src/github.com/thisisaaronland/go-flickr-archive; then mkdir -p src/github.com/thisisaaronland/go-flickr-archive; fi
	cp -r archive src/github.com/thisisaaronland/go-flickr-archive/
	cp -r flickr src/github.com/thisisaaronland/go-flickr-archive/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson/"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt archive/*.go
	go fmt flickr/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/flickr-archive cmd/flickr-archive.go