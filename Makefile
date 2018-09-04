PHONY: go-run
go-run:
	go run linux-namespace.go

PHONY: build
build:
	go build linux-namespace.go

PHONY: run
run: build
	sudo ./linux-namespace

PHONY: clean
clean:
	rm ./linux-namespace