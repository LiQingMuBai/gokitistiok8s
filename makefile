PROJECT_NAME = gokitistiok8s
BINARY_PREFIX = ${PROJECT_NAME}
IMAGE_PREFIX = cage1016/${BINARY_PREFIX}
BUILD_DIR = build
SERVICES = addsvc router foosvc squaresvc
DOCKERS_CLEANBUILD = $(addprefix cleanbuild_docker_,$(SERVICES))
DOCKERS = $(addprefix dev_docker_,$(SERVICES))
DOCKERS_DEBUG = $(addprefix debug_docker_,$(SERVICES))
STAGES = dev debug prod
CGO_ENABLED ?= 0
GOOS ?= linux
COMMIT_HASH = $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP = $(shell date +%Y-%m-%dT%T%z)
BUILD_VERSION ?= 1.0.0
GOLDFLAGS = -s -w
DEBUG_GOGCFLAGS = -gcflags='all=-N -l' -ldflags "$(GOLDFLAGS)"
GOGCFLAGS = -ldflags "$(GOLDFLAGS)"
SHELL  := env BUILD_TAGS=$(BUILD_TAGS) $(SHELL)
BUILD_TAGS ?= "alpha"

define compile_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -tags ${BUILD_TAGS} $(2) -ldflags "-X github.com/cage1016/${PROJECT_NAME}/pkg/$(1)/service.CommitHash=$(COMMIT_HASH) -X github.com/cage1016/${PROJECT_NAME}/pkg/$(1)/service.BuildTimeStamp=$(BUILD_TIMESTAMP) -X github.com/cage1016/${PROJECT_NAME}/pkg/$(1)/service.Version=${BUILD_VERSION}" -o ${BUILD_DIR}/${BINARY_PREFIX}-$(1) cmd/$(1)/main.go
endef

define make_docker_cleanbuild
	docker build --no-cache --build-arg COMMIT_HASH=$(COMMIT_HASH) --build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) --build-arg VERSION=$(BUILD_VERSION) --build-arg PROJECT_NAME=${PROJECT_NAME} --build-arg BINARY=${BINARY_PREFIX}-$(1) --tag=${IMAGE_PREFIX}-$(1) -f deployments/docker/Dockerfile.cleanbuild .
endef

define make_docker
	docker build --build-arg COMMIT_HASH=$(COMMIT_HASH) --build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) --build-arg VERSION=$(BUILD_VERSION) --build-arg BINARY=${BINARY_PREFIX}-$(1) --tag=${IMAGE_PREFIX}-$(1) -f deployments/docker/$(2) ./build
endef

all: $(SERVICES)

.PHONY: all $(SERVICES) dev_dockers debug_dockers cleanbuild_dockers test

cleandocker:
	# Remove gokitistiok8s containers
	docker ps -f name=${IMAGE_PREFIX}-* -aq | xargs docker rm
	# Remove old gokitistiok8s images
	docker images -q ${IMAGE_PREFIX}-* | xargs docker rmi

# Clean ghost docker images
cleanghost:
	# Remove exited containers
	docker ps -f status=dead -f status=exited -aq | xargs docker rm -v
	# Remove unused images
	docker images -f dangling=true -q | xargs docker rmi
	# Remove unused volumes
	docker volume ls -f dangling=true -q | xargs docker volume rm

install:
	cp ${BUILD_DIR}/* $(GOBIN)

test:
	# DEBUG=true bash -c "go test -v github.com/cage1016/gokitistiok8s/<package-name> -run ..."
	go test -v -race -tags test $(shell go list ./... | grep -v 'vendor')

PD_SOURCES:=$(shell find ./pb -type d)
proto:
	@for var in $(PD_SOURCES); do \
		if [ -f "$$var/compile.sh" ]; then \
			cd $$var && ./compile.sh; \
			echo "complie $$var/$$(basename $$var).proto"; \
			cd $(PWD); \
		fi \
	done

# Regenerates OPA data from rego files
HAVE_GO_BINDATA := $(shell command -v go-bindata 2> /dev/null)
generate:
ifndef HAVE_GO_BINDATA
	@echo "requires 'go-bindata' (go get -u github.com/kevinburke/go-bindata/go-bindata)"
	@exit 1 # fail
else
	go generate ./...
endif

$(SERVICES):
	$(call compile_service,$(@),${GOGCFLAGS})

$(DOCKERS_CLEANBUILD):
	$(call make_docker_cleanbuild,$(subst cleanbuild_docker_,,$(@)))

$(DOCKERS):
	@echo BUILD_TAGS=${BUILD_TAGS}

	@if [ "$(filter $(@:dev_docker_%=%), $(SERVICES))" != "" ]; then\
		$(call compile_service,$(subst dev_docker_,,$(@)),${GOGCFLAGS});\
		$(call make_docker,$(subst dev_docker_,,$(@)),Dockerfile);\
		if [ "$(PUSH_IMAGE)" == "true" ]; then \
			docker push ${IMAGE_PREFIX}-$(subst dev_docker_,,$(@)); \
		fi \
	else\
		docker build --tag=${IMAGE_PREFIX}-$(@:dev_docker_%=%) --build-arg COMMIT_HASH=$(COMMIT_HASH) --build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) --build-arg VERSION=$(BUILD_VERSION) -f deployments/docker/Dockerfile.mappingsvc .;\
		if [ "$(PUSH_IMAGE)" == "true" ]; then \
			docker push ${IMAGE_PREFIX}-$(@:dev_docker_%=%); \
		fi \
	fi

$(DOCKERS_DEBUG):
	$(call compile_service,$(subst debug_docker_,,$(@)),${DEBUG_GOGCFLAGS})
	$(call make_docker,$(subst debug_docker_,,$(@)),Dockerfile.debug)

services: $(SERVICES)

dev_dockers: $(DOCKERS)

debug_dockers: $(DOCKERS_DEBUG)

cleanbuild_dockers: $(DOCKERS_CLEANBUILD)

rest_sum:
	curl -X "POST" "http://<your-ip-host>/api/v1/add/sum" -H 'Content-Type: application/json; charset=utf-8' -d '{ "a": 3, "b": 34}'

rest_concat:
	curl -X "POST" "http://<your-ip-host>/api/v1/add/concat" -H 'Content-Type: application/json; charset=utf-8' -d '{ "a": "3", "b": "34"}'

rest_foo:
	curl -X "POST" "http://<your-ip-host>/api/v1/foo/foo" -H 'Content-Type: application/json; charset=utf-8' -d '{"s": "3ddd"}'

grpc_sum:
	grpcurl -plaintext -proto ./pb/addsvc/addsvc.proto -d '{"a": 3, "b":5}' <your-ip-host>:443 pb.Addsvc.Sum

grpc_concat:
	grpcurl -plaintext -proto ./pb/addsvc/addsvc.proto -d '{"a": "3", "b":"5"}' <your-ip-host>:443 pb.Addsvc.Concat

grpc_foo:
	grpcurl -plaintext -proto ./pb/foosvc/foosvc.proto -d '{"s": "foo"}' <your-ip-host>:443 pb.Foosvc.Foo

new_squaresvc:
	./tools/gk n s squaresvc
	sed -i "" 's/Foo(ctx context.Context, s string) (res string, err error)/Square(ctx context.Context, s int64) (res int64, err error)/g' pkg/squaresvc/service/service.go
	./tools/gk init squaresvc
	sed -i "" 's/return res, err/return s * s, err/g' pkg/squaresvc/service/service.go
	./tools/gk add grpc squaresvc
	cd pb/squaresvc && ./compile.sh
	./tools/gk init grpc squaresvc
	./tools/gk new cmd squaresvc

grpc_squaresvc:
	grpcurl -plaintext -proto ./pb/squaresvc/squaresvc.proto -d '{"s": 3}' localhost:8181 pb.Squaresvc.Square

http_squaresvc:
	curl -X POST localhost:8180/square -d '{"s":22}'