.PHONY: all deps clean build
all: deps clean build
deps:
	go mod tidy
clean: 
	rm -rf ./main output.zip

build:
	GOOS=linux GOARCH=amd64 CC=/usr/local/musl/bin/musl-gcc CGO_ENABLED=1  go build -a -ldflags '-linkmode external -extldflags "-static -fPIC"' -o ./main .
	zip output.zip main microblog_template
