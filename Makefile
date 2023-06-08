git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/ipfs-force-community/sophon-messager/version.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

build:
	rm -rf sophon-messager
	go build $(GOFLAGS) -o sophon-messager .

tools:
	rm -rf sophon-messager-tools
	go build -o sophon-messager-tools ./tools/main.go
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
	docker build --build-arg https_proxy=$(BUILD_DOCKER_PROXY) --build-arg BUILD_TARGET=sophon-messager -t sophon-messager .
	docker tag sophon-messager filvenus/sophon-messager:$(TAG)

ifdef PRIVATE_REGISTRY
	docker tag sophon-messager $(PRIVATE_REGISTRY)/filvenus/sophon-messager:$(TAG)
endif

docker-push: docker
	docker push $(PRIVATE_REGISTRY)/filvenus/sophon-messager:$(TAG)
