.PHONE: default
default: clean test build

.PHONY: test
test: 
	go test -v .

.PHONY: build
build:
	go build .

.PHONY: clean
clean:
	rm -f epubtool