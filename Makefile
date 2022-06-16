.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/uts demos/uts.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ns demos/ns.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/cgroup demos/cgroup.go

.PHONY: tools
tools:
	apt install memtester

.PHONY: run
run:
	./my-docker run -ti /bin/sh

.PHONY: test
test:
	memtester 100M 1

.PHONY: run.echo
run.echo:
	./go-linux-kernel run echo hello

