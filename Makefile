SOURCE = $(wildcard *.go)
ALL = $(foreach suffix,win.exe linux osx,\
		build/webhooker-$(suffix))

all: $(ALL)

win.exe = windows
osx = darwin
build/webhooker-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 go build -o $@
