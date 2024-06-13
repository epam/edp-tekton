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
KUBE_VERSION?=1.28
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

HELMDOCS = ${CURRENT_DIR}/bin/helm-docs
.PHONY: helmdocs
helmdocs: ## Download helm-docs locally if necessary.
	$(call go-get-tool,$(HELMDOCS),github.com/norwoodj/helm-docs/cmd/helm-docs,v1.11.0)

GITCHGLOG = ${CURRENT_DIR}/bin/git-chglog
.PHONY: git-chglog
git-chglog: ## Download git-chglog locally if necessary.
	$(call go-get-tool,$(GITCHGLOG),github.com/git-chglog/git-chglog/cmd/git-chglog,v0.15.4)

.PHONY: build
build: clean ## build interceptor binary
	CGO_ENABLED=0 GOOS=${HOST_OS} GOARCH=${HOST_ARCH} go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/edpinterceptor ./cmd/interceptor/main.go

.PHONY: clean
clean:  ## clean up
	-rm -rf ${DIST_DIR}

.PHONY: test ## Run tests
test: test-chart
	go test ./... -coverprofile=coverage.out `go list ./...`

.PHONY: lint
lint: golangci-lint ## Run go lint
	${GOLANGCILINT} run

test-chart: ${CURRENT_DIR}/.venv/bin/activate
	( \
		source $^; \
		pip3 install -r ${CURRENT_DIR}/requirements.txt; \
		helm repo add epamedp https://epam.github.io/edp-helm-charts/stable; \
		helm dependency update ./charts/pipelines-library; \
		helm dependency update ./charts/custom-pipelines; \
		pytest -sv ./charts/pipelines-library/ --color=yes -n auto; \
		pytest -sv ./charts/custom-pipelines/ --color=yes -n auto; \
	)

${CURRENT_DIR}/.venv/bin/activate:
	python3 -m venv ${CURRENT_DIR}/.venv

## Run e2e tests. Requires kind with running cluster and kuttl tool.
e2e: build
	docker build --no-cache -t ${CONTAINER_REGISTRY_URL}/${CONTAINER_REGISTRY_SPACE}/${E2E_IMAGE_REPOSITORY}:${E2E_IMAGE_TAG} .
	kind load --name $(KIND_CLUSTER_NAME) docker-image ${CONTAINER_REGISTRY_URL}/${CONTAINER_REGISTRY_SPACE}/${E2E_IMAGE_REPOSITORY}:${E2E_IMAGE_TAG}
	E2E_IMAGE_REPOSITORY=${E2E_IMAGE_REPOSITORY} CONTAINER_REGISTRY_URL=${CONTAINER_REGISTRY_URL} CONTAINER_REGISTRY_SPACE=${CONTAINER_REGISTRY_SPACE} E2E_IMAGE_TAG=${E2E_IMAGE_TAG} kubectl-kuttl test

GOLANGCILINT = ${CURRENT_DIR}/bin/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-get-tool,$(GOLANGCILINT),github.com/golangci/golangci-lint/cmd/golangci-lint,v1.58.1)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
go get -d $(2)@$(3) ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: start-kind
start-kind:     ## Start kind cluster
ifeq (true,$(START_KIND_CLUSTER))
	kind create cluster --name $(KIND_CLUSTER_NAME) --config $(KIND_CONFIG) --wait 1m
endif
