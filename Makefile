git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/filecoin-project/venus-messager/version.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

build:
	rm -rf venus-messager
	go build $(GOFLAGS) -o venus-messager .

tools:
	rm -rf venus-messager-tools
	go build -o venus-messager-tools ./tools/main.go
.PHONY: tools

gen:
	go generate ./...


lint:
	golangci-lint run

test:
	go test -race ./...


.PHONY: docker

TAG:=test
docker:
ifdef DOCKERFILE
	cp $(DOCKERFILE) ./dockerfile
else
	curl -O https://raw.githubusercontent.com/filecoin-project/venus-docs/master/script/docker/dockerfile
endif
	docker build --build-arg https_proxy=$(BUILD_DOCKER_PROXY) --build-arg BUILD_TARGET=venus-messager -t venus-messager .
	docker tag venus-messager filvenus/venus-messager:$(TAG)

ifdef PRIVATE_REGISTRY
	docker tag venus-messager $(PRIVATE_REGISTRY)/filvenus/venus-messager:$(TAG)
endif

docker-push: docker
	docker push $(PRIVATE_REGISTRY)/filvenus/venus-messager:$(TAG)
