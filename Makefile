PACKAGE=github.com/epam/edp-common/pkg/config
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

VERSION?=$(shell git describe --tags)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
KUBECTL_VERSION=$(shell go list -m all | grep k8s.io/client-go| cut -d' ' -f2)
HOST_OS?=$(shell go env GOOS)
HOST_ARCH?=$(shell go env GOARCH)

# Use kind cluster for testing
START_KIND_CLUSTER?=true
KIND_CLUSTER_NAME?="tekton"
KUBE_VERSION?=1.34
KIND_CONFIG?=./hack/kind-$(KUBE_VERSION).yaml

CONTAINER_REGISTRY_URL?="repo"
CONTAINER_REGISTRY_SPACE?="edp"
E2E_IMAGE_REPOSITORY?="tekton-image"
E2E_IMAGE_TAG?="latest"

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.kubectlVersion=${KUBECTL_VERSION}

ifneq (${GIT_TAG},)
LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

override GCFLAGS +=all=-trimpath=${CURRENT_DIR}

.DEFAULT_GOAL:=help
# set default shell
SHELL=/bin/bash -o pipefail -o errexit
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: validate-docs
validate-docs: helm-docs  ## Validate helm docs
	@git diff -s --exit-code charts/*/README.md || (echo "Run 'make helm-docs' to address the issue." && git diff && exit 1)

# use https://github.com/git-chglog/git-chglog/
.PHONY: changelog
changelog: git-chglog	## generate changelog
ifneq (${NEXT_RELEASE_TAG},)
	$(GITCHGLOG) --next-tag v${NEXT_RELEASE_TAG} -o CHANGELOG.md v0.6.0..
else
	$(GITCHGLOG) -o CHANGELOG.md v0.6.0..
endif

.PHONY: helm-docs
helm-docs: helmdocs	## generate helm docs
	$(HELMDOCS)

.PHONY: build
build: fmt vet ## build interceptor binary
	CGO_ENABLED=0 GOOS=${HOST_OS} GOARCH=${HOST_ARCH} go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/edpinterceptor-${HOST_ARCH} ./cmd/interceptor/main.go

.PHONY: clean
clean:  ## clean up
	-rm -rf ${DIST_DIR}

.PHONY: test ## Run tests
test: test-chart test-go

test-go: ## Run go tests only
	go test ./... -coverprofile=coverage.out `go list ./...`

.PHONY: fmt
fmt:  ## Run go fmt
	go fmt ./...

.PHONY: vet
vet:  ## Run go vet
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run go lint
	$(GOLANGCI_LINT) run -v ./...

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix -v ./...

test-chart: ${CURRENT_DIR}/.venv/bin/activate
	( \
		source $^; \
		pip3 install -r ${CURRENT_DIR}/requirements.txt; \
		helm repo add epamedp https://epam.github.io/edp-helm-charts/stable; \
		helm dependency update ./charts/pipelines-library; \
		pytest -sv ./charts/pipelines-library/ --color=yes -n auto; \
	)

${CURRENT_DIR}/.venv/bin/activate:
	python3 -m venv ${CURRENT_DIR}/.venv

## Run e2e tests. Requires kind with running cluster and kuttl tool.
e2e: build
	docker build --no-cache -t ${CONTAINER_REGISTRY_URL}/${CONTAINER_REGISTRY_SPACE}/${E2E_IMAGE_REPOSITORY}:${E2E_IMAGE_TAG} .
	kind load --name $(KIND_CLUSTER_NAME) docker-image ${CONTAINER_REGISTRY_URL}/${CONTAINER_REGISTRY_SPACE}/${E2E_IMAGE_REPOSITORY}:${E2E_IMAGE_TAG}
	E2E_IMAGE_REPOSITORY=${E2E_IMAGE_REPOSITORY} CONTAINER_REGISTRY_URL=${CONTAINER_REGISTRY_URL} CONTAINER_REGISTRY_SPACE=${CONTAINER_REGISTRY_SPACE} E2E_IMAGE_TAG=${E2E_IMAGE_TAG} kubectl-kuttl test

.PHONY: start-kind
start-kind:     ## Start kind cluster
ifeq (true,$(START_KIND_CLUSTER))
	kind create cluster --name $(KIND_CLUSTER_NAME) --config $(KIND_CONFIG) --wait 1m
endif

## Location to install dependencies to
LOCALBIN ?= ${CURRENT_DIR}/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Tools version
GOLANGCI_LINT_VERSION ?=v2.8.0
HELMDOCS_VERSION ?= v1.14.2
GITCHGLOG_VERSION ?= v0.15.4

HELMDOCS = $(LOCALBIN)/helm-docs
.PHONY: helmdocs
helmdocs: ## Download helm-docs locally if necessary.
	$(call go-install-tool,$(HELMDOCS),github.com/norwoodj/helm-docs/cmd/helm-docs,$(HELMDOCS_VERSION))

GITCHGLOG = $(LOCALBIN)/git-chglog
.PHONY: git-chglog
git-chglog: ## Download git-chglog locally if necessary.
	$(call go-install-tool,$(GITCHGLOG),github.com/git-chglog/git-chglog/cmd/git-chglog,$(GITCHGLOG_VERSION))

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
