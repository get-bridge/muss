BUILD_OS_LIST = darwin linux
BUILD_ARCH_LIST = 386 amd64

MODULE = github.com/get-bridge/muss
VERSION_VAR = $(MODULE)/cmd.Version
VERSION = $(shell git describe --tags --long --always --match 'v[0-9]*' | sed -e 's/-/./')
BUILD_ARGS = -tags netgo -ldflags '-X $(VERSION_VAR)=$(VERSION)'

.PHONY: build

all: test build release

build:
	go build $(BUILD_ARGS) -o build/muss

install:
	go install $(BUILD_ARGS)

release:
	rm -f build/muss build/muss-*.zip build/SHA256SUMS; \
	for os in $(BUILD_OS_LIST); do \
		for arch in $(BUILD_ARCH_LIST); do \
			built=muss zip=muss-$$os-$$arch.zip; \
			GOOS=$$os GOARCH=$$arch go build $(BUILD_ARGS) -o build/$$built \
			&& (cd build && rm -f $$zip && zip $$zip $$built && rm -f $$built); \
		done; \
	done; \
	(cd build && sha256sum *.zip > SHA256SUMS)

# Include "version" as a basic smoke test.
test: version
	# test load order to ensure env vars are respected.
	MUSS_FILE=testdata/muss-env/muss.yaml MUSS_USER_FILE=testdata/muss-env/user.yaml go run . config show --format '{{ yaml user }}' | grep -q 'muss_user_file: respected'
	mkdir -p coverage
	go test -coverprofile=coverage/out -v ./...
	go tool cover -html=coverage/out -o coverage/index.html

version:
	go run $(BUILD_ARGS) . version
