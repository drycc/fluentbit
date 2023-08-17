SHORT_NAME ?= fluentbit
BUILD_TAG ?= git-$(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date --rfc-3339=ns | tr " " T)
SHELL_SCRIPTS = $(wildcard rootfs/usr/local/bin/*)
DRYCC_REGISTRY ?= ${DEV_REGISTRY}
IMAGE_PREFIX ?= drycc
PLATFORM ?= linux/amd64,linux/arm64
REPO_PATH := github.com/drycc/${SHORT_NAME}
DEV_ENV_BUILD = go build -ldflags "-X 'main.Revision=$(BUILD_TAG)' -X 'main.BuildDate=$(BUILD_DATE)'" -buildmode=c-shared -o _dist/out_drycc.so plugin/out_drycc.go
DEV_ENV_IMAGE := ${DEV_REGISTRY}/drycc/go-dev
DEV_ENV_WORK_DIR := /opt/drycc/go/src/${REPO_PATH}
DEV_ENV_PREFIX := docker run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}

include versioning.mk


build: docker-build
push: docker-push

check-docker:
	@if [ -z $$(which docker) ]; then \
	  echo "Missing docker client which is required for development"; \
	  exit 2; \
	fi

bootstrap: check-docker
	$(DEV_ENV_PREFIX) $(DEV_ENV_IMAGE) go mod vendor

build-binary: check-docker
	$(DEV_ENV_PREFIX) $(DEV_ENV_IMAGE) $(DEV_ENV_BUILD)

docker-build: check-docker
	docker build --build-arg CODENAME=${CODENAME} --build-arg BUILD_TAG=${BUILD_TAG} --build-arg BUILD_DATE=${BUILD_DATE} -t ${IMAGE} .
	docker tag ${IMAGE} ${MUTABLE_IMAGE}

docker-buildx: check-docker
	docker buildx build --platform ${PLATFORM} --build-arg CODENAME=${CODENAME} --build-arg BUILD_TAG=${BUILD_TAG} --build-arg BUILD_DATE=${BUILD_DATE} -t ${IMAGE} . --push

test: test-style

test-style: check-docker
	$(DEV_ENV_PREFIX) $(DEV_ENV_IMAGE) make style-check

style-check:
	gofmt -l -w -s plugin
	go vet plugin
	lint
	shellcheck $(SHELL_SCRIPTS)

clean:
	rm -rf _dist
