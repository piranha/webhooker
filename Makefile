SOURCE = $(wildcard *.go)
ALL = $(foreach suffix,win.exe linux osx,\
		build/webhooker-$(suffix))

all: $(ALL)

win.exe = windows
osx = darwin
build/webhooker-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 go build -o $@

upload: $(ALL)
ifndef UPLOAD_PATH
	@echo "Define UPLOAD_PATH to determine where files should be uploaded"
else
	rsync -l -P $(ALL) $(UPLOAD_PATH)
endif

test:
	go test
	cram tests/cram.t
