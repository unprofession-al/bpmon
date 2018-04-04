dest = ./bin

all: clean build

clean:
	rm -rf $(dest)/*

build:
	go generate ./...
	CGO_ENABLED=0 go build -a -installsuffix cgo -o ./bpmon ./cmd/bpmon
