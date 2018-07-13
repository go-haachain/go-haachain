# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: ghaa android ios ghaa-cross swarm evm all test clean
.PHONY: ghaa-linux ghaa-linux-386 ghaa-linux-amd64 ghaa-linux-mips64 ghaa-linux-mips64le
.PHONY: ghaa-linux-arm ghaa-linux-arm-5 ghaa-linux-arm-6 ghaa-linux-arm-7 ghaa-linux-arm64
.PHONY: ghaa-darwin ghaa-darwin-386 ghaa-darwin-amd64
.PHONY: ghaa-windows ghaa-windows-386 ghaa-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

ghaa:
	build/env.sh go run build/ci.go install ./cmd/ghaa
	@echo "Done building."
	@echo "Run \"$(GOBIN)/ghaa\" to launch ghaa."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/ghaa.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Ghaa.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

ghaa-cross: ghaa-linux ghaa-darwin ghaa-windows ghaa-android ghaa-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-*

ghaa-linux: ghaa-linux-386 ghaa-linux-amd64 ghaa-linux-arm ghaa-linux-mips64 ghaa-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-*

ghaa-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/ghaa
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep 386

ghaa-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/ghaa
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep amd64

ghaa-linux-arm: ghaa-linux-arm-5 ghaa-linux-arm-6 ghaa-linux-arm-7 ghaa-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep arm

ghaa-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/ghaa
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep arm-5

ghaa-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/ghaa
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep arm-6

ghaa-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/ghaa
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep arm-7

ghaa-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/ghaa
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep arm64

ghaa-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/ghaa
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep mips

ghaa-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/ghaa
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep mipsle

ghaa-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/ghaa
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep mips64

ghaa-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/ghaa
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-linux-* | grep mips64le

ghaa-darwin: ghaa-darwin-386 ghaa-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-darwin-*

ghaa-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/ghaa
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-darwin-* | grep 386

ghaa-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/ghaa
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-darwin-* | grep amd64

ghaa-windows: ghaa-windows-386 ghaa-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-windows-*

ghaa-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/ghaa
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-windows-* | grep 386

ghaa-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/ghaa
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/ghaa-windows-* | grep amd64
