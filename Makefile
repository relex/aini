SOURCES ?= $(shell find . -name '*.go')
SOURCES_NONTEST ?= $(shell find . -name '*.go' -not -name '*_test.go')

.PHONY: test
test:
	go test -timeout $${TEST_TIMEOUT:-10s} -v ./...

# test-all ignores testcache (go clean testcache)
.PHONY: test-all
test-all:
	go test -timeout $${TEST_TIMEOUT:-10s} -v -count=1 ./...

.PHONY: upgrade
upgrade:
	rm -f go.sum
	go get -u -d ./...; go mod tidy
