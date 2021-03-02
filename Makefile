build:
	rm -rf venus-messager
	go build -o venus-messager .

deps:
	git submodule update --init
	./extern/filecoin-ffi/install-filcrypto

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

