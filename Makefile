git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/filecoin-project/venus-messager/version.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

build: tools
	rm -rf venus-messager
	go build $(GOFLAGS) -o venus-messager .

tools:
	rm -rf venus-messager-tools
	go build -o venus-messager-tools ./tools/main.go
.PHONY: tools

gen:
	go run ./gen/gen.go > ./api/controller/auth_map.go
	gofmt -e -s -w ./api/controller/auth_map.go
.PHONY: gen

deps:
	git submodule update --init
	./extern/filecoin-ffi/install-filcrypto

lint:
	golangci-lint run

test:
	go test -race ./...


.PHONY: docker



docker:
	docker build --build-arg https_proxy=$(BUILD_DOCKER_PROXY) -t venus-messager .
