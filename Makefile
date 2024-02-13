PACKAGE := github.com/marcsello/webploy-server

# define the build timestamp, commit hash, version
BUILD_TIMESTAMP := $(shell date '+%Y-%m-%dT%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_VERSION ?= $(DRONE_TAG)

ifeq ($(BUILD_VERSION),)
	# fallback to the branch name
	BUILD_VERSION := $(shell git rev-parse --abbrev-ref HEAD)
endif

# define build parameters
export GOOS := linux
LDFLAGS := -X 'main.version=${BUILD_VERSION}' -X 'main.commitHash=${COMMIT_HASH}' -X 'main.buildTimestamp=${BUILD_TIMESTAMP}'

.PHONY: all
all: main_amd64 main_arm64 main_arm

# because output file names change dynamically, I'll just make these phony targets
.PHONY: main_amd64
main_amd64: main.go dist
	GOARCH=amd64 go build -v -ldflags="${LDFLAGS}" -o "dist/webploy_${BUILD_VERSION}_${GOOS}_amd64" "."

.PHONY: main_arm64
main_arm64: main.go dist
	GOARCH=arm64 go build -v -ldflags="${LDFLAGS}" -o "dist/webploy_${BUILD_VERSION}_${GOOS}_arm64" "."

.PHONY: main_arm
main_arm: main.go dist
	GOARCH=arm go build -v -ldflags="${LDFLAGS}" -o "dist/webploy_${BUILD_VERSION}_${GOOS}_arm" "."

dist:
	mkdir -p dist

.PHONY: clean
clean:
	rm -rf dist/