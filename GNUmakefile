TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=zcc
# LINT_PKG is the import path the lint target inspects. ZCC's
# provider code lives under ./internal/... (Plugin Framework layout),
# unlike terraform-provider-zia / terraform-provider-zpa which still
# keep SDK v2 code at ./$(PKG_NAME). See the `lint:` target below for
# why we do not use tfproviderlint on Framework code.
LINT_PKG=./internal/...
GOFMT:=gofumpt
TFPROVIDERLINT=tfproviderlint
STATICCHECK=staticcheck
TF_PLUGIN_DIR=~/.terraform.d/plugins
ZCC_PROVIDER_NAMESPACE=zscaler.com/zcc/zcc

# Expression to match against tests
# go test -run <filter>
# e.g. `make testacc TEST_FILTER=TestAccTrustedNetwork_basic` runs that test only.
#
# We compute RUN_FLAG inside an ifdef instead of mutating TEST_FILTER
# itself. GNU Make's rule is that variables set on the command line
# override every plain assignment in the makefile body (even `:=`)
# unless `override` is used — so the older pattern
#
#     ifdef TEST_FILTER
#         TEST_FILTER := -run $(TEST_FILTER)
#     endif
#
# silently dropped the `-run ` prefix when TEST_FILTER was passed on
# the command line, causing go test to see the regex as a positional
# argument and quietly run the entire TestAcc* suite. Splitting into a
# dedicated RUN_FLAG variable avoids that surprise.
ifdef TEST_FILTER
RUN_FLAG := -run $(TEST_FILTER)
endif

TESTARGS?=-test.v

default: build

dep: # Download required dependencies
	go mod tidy

docs:
	go generate

build: fmtcheck
	go install

clean:
	go clean -cache -testcache ./...

clean-all:
	go clean -cache -testcache -modcache ./...

test:
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) $(RUN_FLAG) -timeout=30s -parallel=10

testacc:
	TF_ACC=1 go test $(TEST) $(TESTARGS) $(RUN_FLAG) -timeout 120m

# test:integration:zcc is the CI entry point. It exercises the
# Plugin-Framework code under ./internal/framework/... (this provider
# does not ship SDK v2 resources, so there is no top-level ./zcc
# package as in terraform-provider-zia / terraform-provider-zpa).
# TF_ACC=1 is set explicitly so the recipe works the same locally and
# in CI, and -timeout 120m mirrors the `testacc` cap so live-API tests
# (singletons, ImportStateVerify steps, etc.) cannot be killed by Go's
# default 10-minute per-binary timeout.
test\:integration\:zcc:
	@echo "$(COLOR_ZSCALER)Running zcc integration tests...$(COLOR_NONE)"
	TF_ACC=1 go test -v -race -cover -coverprofile=zcccoverage.out -covermode=atomic ./internal/framework/... -parallel 1 -timeout 120m
	go tool cover -html=zcccoverage.out -o zcccoverage.html
	go tool cover -func zcccoverage.out | grep total:

build13: GOOS=$(shell go env GOOS)
build13: GOARCH=$(shell go env GOARCH)
ifeq ($(OS),Windows_NT)  # is Windows_NT on XP, 2000, 7, Vista, 10...
build13: DESTINATION=$(APPDATA)/terraform.d/plugins/$(ZCC_PROVIDER_NAMESPACE)/0.1.0/$(GOOS)_$(GOARCH)
else
build13: DESTINATION=$(HOME)/.terraform.d/plugins/$(ZCC_PROVIDER_NAMESPACE)/0.1.0/$(GOOS)_$(GOARCH)
endif
build13: fmtcheck
	@echo "==> Installing plugin to $(DESTINATION)"
	@mkdir -p $(DESTINATION)
	go build -o $(DESTINATION)/terraform-provider-zcc_v0.1.0

vet:
	@echo "==> Checking source code against go vet and staticcheck"
	@go vet ./...
	@staticcheck ./...

imports:
	goimports -w $(GOFMT_FILES)

fmt: tools # Format the code
	@echo "formatting the code with $(GOFMT)..."
	@$(GOFMT) -l -w .

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

fmt-docs:
	@echo "✓ Formatting code samples in documentation"
	@terrafmt fmt -p '*.md' .

vendor-status:
	@govendor status

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

lint:
	@echo "==> Checking source code against linters..."
	@# tfproviderlint's -AT001 / -R004 / -S00x checks are SDK v2-specific:
	@# they require *schema.Resource / *schema.Schema composite literals
	@# to analyse and fail their prerequisites on Plugin Framework code.
	@# This provider is Plugin Framework only (no SDK v2 surface), so we
	@# mirror terraform-provider-confluent's stance and rely on tools
	@# that actually understand the Framework AST: gofumpt for format,
	@# go vet for correctness, staticcheck for semantic bugs.
	@echo "==> gofumpt (format check)"
	@if [ -n "$$($(GOFMT) -l $(LINT_PKG:./%/...=./%) 2>/dev/null)" ]; then \
		echo "gofumpt findings (run 'make fmt' to fix):"; \
		$(GOFMT) -l $(LINT_PKG:./%/...=./%); \
		exit 1; \
	fi
	@echo "==> go vet"
	@go vet $(LINT_PKG)
	@echo "==> staticcheck"
	@$(STATICCHECK) $(LINT_PKG)

tools:
	@which $(GOFMT) || go install mvdan.cc/gofumpt@v0.9.2
	@which $(TFPROVIDERLINT) || go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
	@which $(STATICCHECK) || go install honnef.co/go/tools/cmd/staticcheck@latest

tools-update:
	@go install mvdan.cc/gofumpt@v0.9.2
	@go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build test testacc vet fmt fmtcheck errcheck tools vendor-status test-compile website-lint website website-test

