GO ?= go

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")
PROJECT := github.com/opencontainers-sec/go-containersec

COMMIT ?= $(shell git describe --dirty --long --always)

GOARCH := $(shell $(GO) env GOARCH)

GO_BUILDMODE :=
# Enable dynamic PIE executables on supported platforms.
ifneq (,$(filter $(GOARCH),386 amd64 arm arm64 ppc64le riscv64 s390x))
	ifeq (,$(findstring -race,$(EXTRA_FLAGS)))
		GO_BUILDMODE := "-buildmode=pie"
	endif
endif
GO_BUILD := $(GO) build -trimpath $(GO_BUILDMODE) \
	$(EXTRA_FLAGS) \
	-ldflags "$(LDFLAGS_COMMON) $(EXTRA_LDFLAGS)"

GO_BUILDMODE_STATIC :=
LDFLAGS_STATIC := -extldflags -static
# Enable static PIE executables on supported platforms.
# This (among the other things) requires libc support (rcrt1.o), which seems
# to be available only for arm64 and amd64 (Debian Bullseye).
ifneq (,$(filter $(GOARCH),arm64 amd64))
	ifeq (,$(findstring -race,$(EXTRA_FLAGS)))
		GO_BUILDMODE_STATIC := -buildmode=pie
		LDFLAGS_STATIC := -linkmode external -extldflags -static-pie
	endif
endif
# Enable static PIE binaries on supported platforms.
GO_BUILD_STATIC := $(GO) build -trimpath $(GO_BUILDMODE_STATIC) \
	$(EXTRA_FLAGS) \
	-ldflags "$(LDFLAGS_COMMON) $(LDFLAGS_STATIC) $(EXTRA_LDFLAGS)"

.DEFAULT: dmz

.PHONY: dmz
dmz:
	$(GO_BUILD_STATIC) -o dmz cmd/dmz.go

.PHONY: clean
clean:
	rm -rf ./dmz