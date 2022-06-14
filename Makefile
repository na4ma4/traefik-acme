DOCKER_REPO := ghcr.io/na4ma4/traefik-acme

GO_MATRIX += darwin/amd64 darwin/arm64
GO_MATRIX += linux/amd64
GO_MATRIX += windows/amd64

DOCKER_PLATFORMS := linux/$(shell go env GOARCH)
DOCKER_BUILD_ARGS += --build-arg "DOCKER_PLATFORM=$(DOCKER_PLATFORMS)"

APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH ?= $(shell git show -s --format=%h)

GO_DEBUG_ARGS   ?= -v -ldflags "-X main.version=$(GO_APP_VERSION)+debug -X main.commit=$(GIT_HASH) -X main.date=$(APP_DATE) -X main.builtBy=makefiles"
GO_RELEASE_ARGS ?= -v -ldflags "-X main.version=$(GO_APP_VERSION) -X main.commit=$(GIT_HASH) -X main.date=$(APP_DATE) -X main.builtBy=makefiles -s -w"

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/pkg/docker/v1/Makefile
-include .makefiles/ext/na4ma4/lib/golangci-lint/v1/Makefile
-include .makefiles/ext/na4ma4/lib/goreleaser/v1/Makefile

.makefiles/ext/na4ma4/%: .makefiles/Makefile
	@curl -sfL https://raw.githubusercontent.com/na4ma4/makefiles-ext/main/v1/install | bash /dev/stdin "$@"

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"

.PHONY: install
install: artifacts/build/release/$(GOHOSTOS)/$(GOHOSTARCH)/traefik-acme
	install "$(<)" /usr/local/bin/

.PHONY: run
run: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/traefik-acme
	$< $(RUN_ARGS)

.PHONY: upx
upx: $(patsubst artifacts/build/%,artifacts/upx/%.upx,$(_GO_RELEASE_TARGETS_ALL))

artifacts/upx/%.upx: artifacts/build/%
	-@mkdir -p "$(@D)"
	-$(RM) -f "$(@)"
	upx -o "$@" "$<"

.PHONY: docker-run
docker-run: docker
	docker run -ti --rm $(DOCKER_REPO):$(DOCKER_TAGS) $(RUN_ARGS)


######################
# Tests
######################

TEST_RUNNER := artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/traefik-acme

.DELETE_ON_ERROR: artifacts/test/issue-5/test1.out
REGRESSION_TESTS += artifacts/test/issue-5/test1.out
artifacts/test/issue-5/test1.out: test/issue-5/acme.json $(TEST_RUNNER) $(GO_SOURCE_FILES)
	-@mkdir -p "$(@D)"
	-@$(RM) "$(@D)/cert.pem" "$(@D)/key.pem"
	$(TEST_RUNNER) --acme="$(<)" --certificate-resolver="acme-different" --cert "$(@D)/cert.pem" --key "$(@D)/key.pem" "*.example.com" | tee "$(@)"
	grep "certificate-for-example.com" "$(@D)/cert.pem"
	grep "key-for-example.com" "$(@D)/key.pem"

.DELETE_ON_ERROR: artifacts/test/issue-5/test2.out
REGRESSION_TESTS += artifacts/test/issue-5/test2.out
artifacts/test/issue-5/test2.out: test/issue-5/acme.json $(TEST_RUNNER) $(GO_SOURCE_FILES)
	-@mkdir -p "$(@D)"
	-@$(RM) "$(@D)/cert.pem" "$(@D)/key.pem"
	$(TEST_RUNNER) --acme="$(<)" --certificate-resolver="acme-different" --cert "$(@D)/cert.pem" --key "$(@D)/key.pem" "example.com" | tee "$(@)"
	grep "certificate-for-example.com" "$(@D)/cert.pem"
	grep "key-for-example.com" "$(@D)/key.pem"

.DELETE_ON_ERROR: artifacts/test/issue-14/v1/test1.out
REGRESSION_TESTS += artifacts/test/issue-14/v1/test1.out
artifacts/test/issue-14/v1/test1.out: test/issue-14/v1/acme.json $(TEST_RUNNER) $(GO_SOURCE_FILES)
	-@mkdir -p "$(@D)"
	-@$(RM) "$(@D)/cert.pem" "$(@D)/key.pem"
	$(TEST_RUNNER) --acme="$(<)" --cert "$(@D)/cert.pem" --key "$(@D)/key.pem" "test.example.com" | tee "$(@)"
	grep "Certificate" "$(@D)/cert.pem"
	grep "Certificate Key" "$(@D)/key.pem"

.DELETE_ON_ERROR: artifacts/test/issue-14/v2/test1.out
REGRESSION_TESTS += artifacts/test/issue-14/v2/test1.out
artifacts/test/issue-14/v2/test1.out: test/issue-14/v2/new-acme.json $(TEST_RUNNER) $(GO_SOURCE_FILES)
	-@mkdir -p "$(@D)"
	-@$(RM) "$(@D)/cert.pem" "$(@D)/key.pem"
	$(TEST_RUNNER) --acme="$(<)" --certificate-resolver "myresolver" --cert "$(@D)/cert.pem" --key "$(@D)/key.pem" "test.example.com" | tee "$(@)"
	grep "Certificate" "$(@D)/cert.pem"
	grep "Certificate Key" "$(@D)/key.pem"

.DELETE_ON_ERROR: artifacts/test/issue-52/cert.pem artifacts/test/issue-52/key.pem
REGRESSION_TESTS += artifacts/test/issue-52/cert.pem artifacts/test/issue-52/key.pem
artifacts/test/issue-52/cert.pem artifacts/test/issue-52/key.pem: test/issue-52/acme.json docker
	-@mkdir -p "$(@D)"
	-@$(RM) "$(@D)/cert.pem" "$(@D)/key.pem"
	docker run --rm --user "$(shell id -u):$(shell id -g)" -v "$(shell pwd)/test/issue-52:/input" -v "$(shell pwd)/$(@D):/output" --workdir /output $(DOCKER_REPO):$(DOCKER_TAGS) --acme "/input/acme.json" test.example.com | tee "$(@).log"
	grep "^Certificate" "$(@)"

.PHONY: regression-tests
regression-tests: $(REGRESSION_TESTS)

ci:: $(REGRESSION_TESTS)

######################
# Linting
######################

ci:: lint
