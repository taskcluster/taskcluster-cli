# The name of the executable.
BINARY = taskcluster

# Flags that are to be passed to the linker, can be overwritten by
# the environment or as an argument to make.
LDFLAGS ?=

SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

VERSION := $(shell git describe --always  --dirty --tags)
LDFLAGS += -X github.com/taskcluster/taskcluster-cli/cmds/version.VersionNumber=$(VERSION)

BUILD_ARCH ?= amd64 386
BUILD_OS ?= !netbsd !plan9

# Removing openbsd/386 because gopsutil can't cross compile for it yet.
# Removing darwin/386 until https://github.com/shirou/gopsutil/issues/348 is fixed
BUILD_OSARCH = !openbsd/386 !darwin/386

all: prep build

prep:
	go get github.com/kardianos/govendor
	govendor sync

build: $(BINARY)

$(BINARY): $(SOURCES)
	go build -ldflags "${LDFLAGS}" -o ${BINARY} .

_upload_release/upload: _upload_release/upload.go
	go get ./_upload_release
	go build -o $@ ./_upload_release

release: $(SOURCES)
	go get -u github.com/mitchellh/gox
	rm -rf build/
	gox -os="${BUILD_OS}" -arch="${BUILD_ARCH}" -osarch="${BUILD_OSARCH}" -ldflags "${LDFLAGS}" -output="build/{{.OS}}-{{.Arch}}/${BINARY}"

clean:
	rm -f ${BINARY}
	rm -rf build

test: prep build
	go test -v -race ./...

generate-apis:
	go get github.com/taskcluster/go-got
	go generate ./apis

lint: prep
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	# not enabled: aligncheck, deadcode, dupl, errcheck, gas, gocyclo, structcheck, unused, varcheck
	# Disabled: testify, test (these two show test errors, hence, they run tests)
	# Disabled: gotype (same as go compiler, also it has issues and was recently removed)
	gometalinter -j4 --deadline=30m --line-length=180 --vendor --vendored-linters --disable-all \
		--enable=goconst \
		--enable=gofmt \
		--enable=goimports \
		--enable=golint \
		--enable=gosimple \
		--enable=ineffassign \
		--enable=interfacer \
		--enable=lll \
		--enable=misspell \
		--enable=staticcheck \
		--enable=unconvert \
		--enable=vet \
		--enable=vetshadow \
		--tests ./...

.PHONY: all prep build clean upload release generate-apis
