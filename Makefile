BUILD_OS_LIST = darwin linux
BUILD_ARCH_LIST = 386 amd64

.PHONY: build

all: test build

build:
	for os in $(BUILD_OS_LIST); do \
		for arch in $(BUILD_ARCH_LIST); do \
			built=muss zip=muss-$$os-$$arch.zip; \
			GOOS=$$os GOARCH=$$arch go build -o build/$$built \
			&& (cd build && rm -f $$zip && zip $$zip $$built && rm -f $$built); \
		done; \
	done; \
	(cd build && sha256sum *.zip > SHA256SUMS)

test:
	go test -v ./...
