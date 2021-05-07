git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/filecoin-project/venus-messager/version.GitCommit=${git}
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

build:
	rm -rf venus-messager
	go build $(GOFLAGS) -o venus-messager .
	./venus-messager --version

gen:
	go run ./gen/gen.go > ./api/controller/auth_map.go
	gofmt -e -s -w ./api/controller/auth_map.go
.PHONY: gen

deps:
	git submodule update --init
	./extern/filecoin-ffi/install-filcrypto

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

test:
	rm -rf models/test_sqlite_db*
	go test -race ./...

