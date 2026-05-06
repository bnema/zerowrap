GOLANGCI_LINT_VERSION ?= latest
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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCAL_BIN) $(GOLANGCI_LINT_VERSION)
