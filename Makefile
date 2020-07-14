SOURCE = $(wildcard *.go)
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'

ALL = $(foreach suffix,win.exe linux osx,\
		build/webhooker-$(suffix))

all: $(ALL)

clean:
	rm -f $(ALL)

test:
	go test
	cram tests/cram.t

win.exe = windows
osx = darwin
build/webhooker-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 go build -o $@

ifndef desc
release:
	@echo "Push a tag and run this as 'make release desc=tralala'"
else
release: $(ALL)
	github-release release -u piranha -r webhooker -t "$(TAG)" -n "$(TAG)" --description "$(desc)"
	@for x in $(ALL); do \
		echo "Uploading $$x" && \
		github-release upload -u piranha \
                              -r webhooker \
                              -t $(TAG) \
                              -f "$$x" \
                              -n "$$(basename $$x)"; \
	done
endif
