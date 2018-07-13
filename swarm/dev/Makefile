.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/haachain/go-haachain/cmd/swarm
	go build -o bin/ghaa     github.com/haachain/go-haachain/cmd/ghaa
	go build -o bin/bootnode github.com/haachain/go-haachain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/haachain/go-haachain/swarm/...
