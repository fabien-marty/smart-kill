FIX ?= 1

all: build

.PHONY: build
build: smart-kill

smart-kill: main.go
	go build -o $@ $<

.PHONY: clean
clean:
	rm -f smart-kill golangci-lint

.PHONY: lint
lint: golangci-lint
	go vet
	if test "$${FIX}" = "1"; then \
		gofmt -s .; \
	else \
		gofmt -s -d .; \
	fi
	if test "$${FIX}" = "1"; then \
		./golangci-lint run; \
	else \
		./golangci-lint run --fix; \
	fi

golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b . v1.54.2
	chmod +x $@