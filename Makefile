.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/mydocker
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/uts demos/uts.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ns demos/ns.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/cgroup demos/cgroup.go

.PHONY: tools
tools:
	apt install memtester

.PHONY: run
run:
	./bin/my-docker run -ti sh

.PHONY: run.stress
run.stress:
	./bin/mydocker run -ti -m 100m stress --vm-bytes 50m --vm-keep --vm 1

.PHONY: test
test:
	memtester 100M 1


