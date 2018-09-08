PHONY: go-run
go-run:
	go run linux-namespace.go

PHONY: build
build:
	go build *.go

PHONY: build-mount
build-mount: build
	rm -Rf /tmp/ns-process
	mkdir -p /tmp/ns-process/rootfs
	tar -C /tmp/ns-process/rootfs -xf assets/busybox.tar

# NB: Requires 'sudo' if no new User namespace is created.
PHONY: run
run: build-mount
	./linux-namespace

PHONY: clean
clean:
	rm ./linux-namespace
	rm -Rf /tmp/ns-process