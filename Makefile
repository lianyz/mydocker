.DEFAULT: all

.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/mydocker
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/uts demos/uts.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ns demos/ns.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/cgroup demos/cgroup.go

.PHONY: mount
mount:
	mount -t aufs -o dirs=./image/writeLayer:./image/busybox mydocker ./image/mnt


.PHONY: umount
umount:
	umount mydocker

.PHONY: tools
tools:
	apt install memtester

.PHONY: run
run:
	./bin/mydocker run -ti -ch sh

.PHONY: run.v
run.v:
	./bin/mydocker run -ti -v /root/volume:/containerVolume sh

.PHONY: run.stress
run.stress:
	./bin/mydocker run -ti -m 100m stress --vm-bytes 50m --vm-keep --vm 1

.PHONY: test
test:
	memtester 100M 1


