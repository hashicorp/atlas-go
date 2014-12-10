VERSION = `cat ./main.go | grep "Version = " | cut -d" " -f4 | sed 's/[^"]*"\([^"]*\).*/\1/'`
DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps build

deps:
	go get -d -v ./...
	echo $(DEPS) | xargs -n1 go get -d

build:
	@mkdir -p bin/
	go build -o bin/atlas-go ./v1

test: deps
	go list ./... | xargs -n1 go test -timeout=3s

package: xcompile
	$(eval FILES := $(shell ls build))
	@mkdir -p build/tgz
	for f in $(FILES); do \
		(cd $(shell pwd) && tar -zcvf tgz/$$f.tar.gz $$f); \
		echo $$f; \
	done

.PHONY: all deps build test xcompile package
