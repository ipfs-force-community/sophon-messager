ldflags=-X=github.com/ipfs-force-community/venus-messager/version.GitCommit=`git log -n 1 --format=%H`
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

build:
	rm -rf venus-messager
	go build $(GOFLAGS) -o venus-messager .
	./venus-messager version

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

