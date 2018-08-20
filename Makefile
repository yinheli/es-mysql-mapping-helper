CMD = es-mysql-mapping-helper
GIT_TAG := $(shell (git describe --abbrev=0 --tags 2> /dev/null || echo $${GIT_TAG:-'v0.0.0'}) | head -n1)
GIT_HASH := $(shell (git show-ref --head --hash=8 2> /dev/null || echo $${GIT_HASH:-'00000000'}) | head -n1)
SRC_DIR := $(shell ls -d */|grep -vE 'vendor')

PLATFORMS := linux/386 linux/amd64 linux/arm linux/arm64 darwin/amd64
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
TARGET = release/$(CMD)-$(os)-$(arch)

.PHONY: all
all: clean $(CMD)

.PHONY: fmt
fmt:
	# gofmt code
	gofmt -s -l -w $(SRC_DIR) *.go
	# go tool vet $(SRC_DIR)
	go tool vet *.go

.PHONY: install
install:
	go install \
	-ldflags='-X "main.Build=$(GIT_TAG)-$(GIT_HASH)"' \
	./cmd/$(CMD)

PHONY: $(CMD)
$(CMD):
	go build \
	-o release/$(CMD) \
	-ldflags='-s -w -X "main.Build=$(GIT_TAG)-$(GIT_HASH)"' \
	.

PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build \
	-o $(TARGET)/$(CMD) \
	-ldflags='-s -w -X "main.Build=$(GIT_TAG)-$(GIT_HASH)"' \
	.
	@tar -czf $(TARGET).tar.gz -C $(TARGET) .
	@rm -rf $(TARGET)

.PHONY: pack-all
pack-all: clean $(PLATFORMS)

.PHONY: test
test:
	go test -v -coverprofile .cover.out ./...
	@go tool cover -func=.cover.out
	@go tool cover -html=.cover.out -o .cover.html


.PHONY: clean
clean:
	@rm -rf release

