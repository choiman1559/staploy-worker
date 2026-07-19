APP_VERSION := 1.0.0
export APP_VERSION

MODULE := $(shell grep ^module go.mod | awk '{print $$2}')
SRC_DIR := protobuf
BUILD_OUT_DIR := out
PROTO_OUT_DIR := app/proto
CORE_NAMES := protocol cpus app
CLI_TOOL := /home/$(USER)/Documents/Staploy/staploy-cli/out/temp/go_build_staploy_cli

GOPATH_DEFAULT := $(shell go env GOPATH)
ifeq ($(GOPATH_DEFAULT),)
    GOPATH_DEFAULT := $(HOME)/go
endif
GOPATH_BIN := $(GOPATH_DEFAULT)/bin

TARGET_PATTERNS := $(addprefix %/, $(addsuffix .proto, $(CORE_NAMES)))
ALL_PROTO_FILES := $(wildcard $(SRC_DIR)/*.proto)
PROTO_FILES     := $(filter $(TARGET_PATTERNS), $(ALL_PROTO_FILES))
M_ARGS := $(foreach file,$(PROTO_FILES),--go_opt=M$(subst $(SRC_DIR)/,,$(file))=$(MODULE)/$(PROTO_OUT_DIR))

.PHONY: all proto clean cleanBuild buildBinaries buildPkg $(ARCHES)

all: proto

proto:
	@mkdir -p $(PROTO_OUT_DIR)
	protoc -I=$(SRC_DIR) \
		--plugin=protoc-gen-go=$(GOPATH_BIN)/protoc-gen-go \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		$(M_ARGS) \
		$(PROTO_FILES)

ARCHES := 386 amd64 arm arm64 riscv64 mipsle mips64le
buildBinaries: $(ARCHES)
buildPkg: buildBinaries createPkg

$(ARCHES):
	CGO_ENABLED=0 GOOS=linux GOARCH=$@ go build -ldflags="-s -w -X staploy-worker/app/consts.VERSION=$(APP_VERSION)" -o $(BUILD_OUT_DIR)/$@/staploy staploy-worker

createPkg:
	$(CLI_TOOL) file -f build_pkg.hcl -v

clean:
	rm -rf $(PROTO_OUT_DIR)

cleanBuild:
	rm -rf $(BUILD_OUT_DIR)