build:
	rm -rf venus-messager
	go build -o venus-messager .

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

