dest = ./bin

all: clean build docker

clean:
	rm -rf $(dest)/*

build:
	go generate ./...
	CGO_ENABLED=0 go build -a -installsuffix cgo -o ./bin/bpmon ./cmd/bpmon

docker:
	docker build . -t bpmon

