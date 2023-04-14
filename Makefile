SOURCE = $(wildcard *.go) go.mod go.sum
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags "-w -X main.Version=${TAG}"

test:
	go test
	prysk tests/cmdline.t

clean:
	rm -f $(ALL)

win.exe.gz = GOOS=windows GOARCH=amd64
linux.gz = GOOS=linux GOARCH=amd64
mac-amd64.gz = GOOS=darwin GOARCH=amd64
mac-arm64.gz = GOOS=darwin GOARCH=arm64
build/webhooker-%.gz: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 $($*) $(GOBUILD) -o $(basename $@)
	gzip $(basename $@)

ALL = $(foreach suffix,mac-arm64 mac-amd64 linux win.exe,\
		build/webhooker-$(suffix).gz)
all: $(ALL)
