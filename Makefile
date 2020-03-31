SHELL = /bin/bash

PROJECT_NAME = joyme123/lazykube
TARGET = lazykube
MAJOR_VERSION   = $(shell cat VERSION)
GIT_VERSION = $(shell git log -1 --pretty=format:%h)

IMAGE_NAME = lazykube
REPOSITORY = joyme/${IMAGE_NAME}

BUILD_IMAGE = golang:1.13

build:
	docker run -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make build-local

build-local:
	@rm -rf build
	GOPROXY=https://goproxy.cn GOMODULE111=on CGO_ENABLED=0 \
	go build \
	-ldflags "-B 0x$(shell head -c20 /dev/urandom|od -An -tx1|tr -d ' \n') -X main.Version=${MAJOR_VERSION}(${GIT_VERSION})" \
	-v -o build/${TARGET} \
	./cmd/lazykube
	chmod -R 777 build

image:
	docker build --no-cache --rm -t ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} .
	docker tag ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} ${REPOSITORY}:${MAJOR_VERSION}-${GIT_VERSION}

imagePush:
	docker push ${REPOSITORY}:${MAJOR_VERSION}-${GIT_VERSION}

unittest:
	bash script/unit-test.sh

.PHONY: build