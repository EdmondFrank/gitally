# Top-level Makefile for Gitaly
#
# Responsibilities of this file:
# - create GOPATH in _build with symlink to current dir
# - re-generate _build/Makefile from makegen.go on each run
# - dispatch commands to _build/Makefile
#
# "Magic" should happen in the makegen.go dynamic template. We want
# _build/Makefile to be as static as possible.

BUILD_DIR = _build
PKG = gitlab.com/gitlab-org/gitaly
MAKEGEN = $(BUILD_DIR)/makegen

# These variables are handed down to make in _build
export GOPATH := $(CURDIR)/$(BUILD_DIR)
export PATH := $(PATH):$(GOPATH)/bin
export TEST_REPO_STORAGE_PATH := $(CURDIR)/internal/testhelper/testdata/data

all: build

.PHONY: build
build: prepare-build
	cd $(BUILD_DIR) && make install INSTALL_DEST_DIR=$(CURDIR) 

.PHONY: install
install: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: assemble
assemble: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: binaries
binaries: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: test
test: prepare-build
	cd $(BUILD_DIR) && make $@
	
.PHONY: rspec
rspec: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: verify
verify: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: format
format: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: cover
cover: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: notice
notice: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: race-go
race-go: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: docker
docker: prepare-build
	cd $(BUILD_DIR) && make $@

.PHONY: prepare-build
prepare-build: $(BUILD_DIR)/.ok update-makefile
$(BUILD_DIR)/.ok:
	mkdir -p $(BUILD_DIR)/src/$(shell dirname $(PKG))
	cd $(BUILD_DIR)/src/$(shell dirname $(PKG)) && rm -f $(shell basename $(PKG)) && \
		ln -sf ../../../.. $(shell basename $(PKG))
	touch $@

.PHONY: update-makefile
update-makefile: _build/makegen $(BUILD_DIR)/.ok
	cd $(BUILD_DIR) && ./makegen > Makefile

_build/makegen: _support/makegen.go $(BUILD_DIR)/.ok
	go build -o $@ _support/makegen.go

clean:
	rm -rf $(BUILD_DIR) .ruby-bundle $(TEST_REPO_STORAGE_PATH)
