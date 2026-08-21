GOLANGCI_LINT_VERSION ?= v2.13.1
LOCAL_BIN ?= $(CURDIR)/.bin
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)
GO_TEST_FLAGS ?= -race -v ./...

ifeq ($(GOLANGCI_LINT),)
GOLANGCI_LINT := $(LOCAL_BIN)/golangci-lint
endif

.PHONY: lint test fmt tidy clean

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

test:
	go test $(GO_TEST_FLAGS)

fmt:
	gofmt -w .

tidy:
	go mod tidy

clean:
	rm -rf $(LOCAL_BIN) coverage.out

$(LOCAL_BIN)/golangci-lint:
	mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
