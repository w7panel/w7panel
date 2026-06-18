
CHART_DIR ?= charts/w7panel
CHART_PACKAGE_DIR ?= charts
CHART_NAME ?= $(shell awk '/^name:/ {print $$2; exit}' $(CHART_DIR)/Chart.yaml)
CHART_IMAGE_REPOSITORY ?= $(shell awk '/^[[:space:]]*repository:/ {print $$2; exit}' $(CHART_DIR)/values.yaml)
CHART_IMAGE_TAG ?= $(shell awk '/^[[:space:]]*tag:/ {gsub(/"/, "", $$2); print $$2; exit}' $(CHART_DIR)/values.yaml)
TAG ?=
IMAGE_REPOSITORY ?= $(CHART_IMAGE_REPOSITORY)
IMAGE_TAG ?= $(if $(TAG),$(TAG),$(CHART_IMAGE_TAG))
HELM_PACKAGE_IMAGE_REPOSITORY ?= $(IMAGE_REPOSITORY)
HELM_PACKAGE_IMAGE_TAG ?= $(IMAGE_TAG)
HELM_CHART_VERSION ?= $(patsubst v%,%,$(IMAGE_TAG))
HELM_APP_VERSION ?= $(IMAGE_TAG)
CHART_PACKAGE_VERSION ?= $(HELM_CHART_VERSION)
PLATFORMS ?= linux/amd64,linux/arm64
TEST_PLATFORM ?= linux/amd64
DOCKERFILE ?= Dockerfile

.PHONY: help package-chart publish build-image build-image-test

help:
	@echo "Usage:"
	@echo "  make <target> [VARIABLE=value]"
	@echo ""
	@echo "Targets:"
	@echo "  package-chart      Package $(CHART_DIR) into $(CHART_PACKAGE_DIR)"
	@echo "  build-image        Build and push the Docker image"
	@echo "  build-image-test   Build a local test image without pushing"
	@echo "  publish            Build and push the Docker image, then package the chart"
	@echo ""
	@echo "Variables:"
	@echo "  CHART_DIR          Chart directory, default: $(CHART_DIR)"
	@echo "  CHART_PACKAGE_DIR  Chart package output directory, default: $(CHART_PACKAGE_DIR)"
	@echo "  CHART_PACKAGE_VERSION Helm chart package version, default: $(CHART_PACKAGE_VERSION)"
	@echo "  IMAGE_REPOSITORY   Image repository, default: $(IMAGE_REPOSITORY)"
	@echo "  IMAGE_TAG          Image tag, default: $(IMAGE_TAG)"
	@echo "  HELM_PACKAGE_IMAGE_REPOSITORY Image repository written into Helm package, default: $(HELM_PACKAGE_IMAGE_REPOSITORY)"
	@echo "  HELM_PACKAGE_IMAGE_TAG Image tag written into Helm package, default: $(HELM_PACKAGE_IMAGE_TAG)"
	@echo "  HELM_CHART_VERSION Helm chart version alias, default: $(HELM_CHART_VERSION)"
	@echo "  HELM_APP_VERSION   Helm app version, default: $(HELM_APP_VERSION)"
	@echo "  TAG                Image tag alias, default: empty"
	@echo "  PLATFORMS          Push build platforms, default: $(PLATFORMS)"
	@echo "  TEST_PLATFORM      Local test build platform, default: $(TEST_PLATFORM)"
	@echo "  DOCKERFILE         Dockerfile path, default: $(DOCKERFILE)"
	@echo ""
	@echo "Examples:"
	@echo "  make build-image TAG=v1.2.3"
	@echo "  make build-image IMAGE_REPOSITORY=ghcr.io/org/w7panel IMAGE_TAG=v1.2.3"
	@echo "  make build-image-test"

publish: build-image package-chart

package-chart:
	@tmp_dir=$$(mktemp -d); \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	cp -R "$(CHART_DIR)" "$$tmp_dir/$(CHART_NAME)"; \
	HELM_PACKAGE_IMAGE_REPOSITORY="$(HELM_PACKAGE_IMAGE_REPOSITORY)" HELM_PACKAGE_IMAGE_TAG="$(HELM_PACKAGE_IMAGE_TAG)" CHART_PACKAGE_VERSION="$(CHART_PACKAGE_VERSION)" HELM_APP_VERSION="$(HELM_APP_VERSION)" perl -0pi -e 's/(^version:\s*).*/$${1}$$ENV{CHART_PACKAGE_VERSION}/m; s/(^appVersion:\s*).*/$${1}$$ENV{HELM_APP_VERSION}/m; s/(^image:\n(?:^[ \t].*\n)*?^[ \t]+repository:\s*).*/$${1}$$ENV{HELM_PACKAGE_IMAGE_REPOSITORY}/m; s/(^image:\n(?:^[ \t].*\n)*?^[ \t]+tag:\s*).*/$${1}"$$ENV{HELM_PACKAGE_IMAGE_TAG}"/m' "$$tmp_dir/$(CHART_NAME)/Chart.yaml" "$$tmp_dir/$(CHART_NAME)/values.yaml"; \
	helm package "$$tmp_dir/$(CHART_NAME)" --destination "$(CHART_PACKAGE_DIR)" --version "$(CHART_PACKAGE_VERSION)" --app-version "$(HELM_APP_VERSION)"

build-image:
	docker buildx build --platform $(PLATFORMS) -f $(DOCKERFILE) -t $(IMAGE_REPOSITORY):$(IMAGE_TAG) --push .

build-image-test:
	docker buildx build --platform $(TEST_PLATFORM) -f $(DOCKERFILE) -t $(IMAGE_REPOSITORY):$(IMAGE_TAG)-test --load .
