PHONY: go-run
go-run:
	go run linux-namespace.go

PHONY: build
build:
	go build linux-namespace.go

# NB: Requires 'sudo' if no new User namespace is created.
PHONY: run
run: build
	./linux-namespace

PHONY: clean
clean:
	rm ./linux-namespace