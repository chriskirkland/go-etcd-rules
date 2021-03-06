GO111MODULE := on
export
LINT_VERSION="1.21.0"

.PHONY: all
all: deps lint test

.PHONY: deps
deps:
	go get github.com/mattn/goveralls
	@if ! which golangci-lint >/dev/null || [[ "$$(golangci-lint --version)" != *${LINT_VERSION}* ]]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v${LINT_VERSION}; \
	fi

.PHONY: update-deps
update-deps:
	# Glide is still updated to help migrate to Go Mod
	glide cc
	glide update -v
	go get -u ./...

.PHONY: lint-fix
lint-fix: deps
	golangci-lint run --fix  # Attempts to fix some lint errors

.PHONY: lint
lint: deps
	golangci-lint run

.PHONY: test
test: int-setup
	go test -race -covermode=atomic -coverprofile=coverage.out ./rules/...
	go run v3enginetest/main.go

.PHONY: int-setup
int-setup: int-teardown
	docker run -d -p 2379:2379 --name etcd quay.io/coreos/etcd:v3.2.9 \
		/usr/local/bin/etcd --listen-client-urls http://0.0.0.0:2379 \
		--advertise-client-urls http://0.0.0.0:2379

.PHONY: int-teardown
int-teardown:
	docker rm -f etcd || true
