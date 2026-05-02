MODULE := $(shell grep ^module go.mod | awk '{print $$2}')
SRC_DIR := protobuf
OUT_DIR := proto

GOPATH_DEFAULT := $(shell go env GOPATH)
ifeq ($(GOPATH_DEFAULT),)
    GOPATH_DEFAULT := $(HOME)/go
endif
GOPATH_BIN := $(GOPATH_DEFAULT)/bin
PROTO_FILES := $(shell find $(SRC_DIR) -name "*.proto")
M_ARGS := $(foreach file,$(PROTO_FILES),--go_opt=M$(subst $(SRC_DIR)/,,$(file))=$(MODULE)/$(OUT_DIR))

.PHONY: all proto clean

all: proto

proto:
	@mkdir -p $(OUT_DIR)
	protoc -I=$(SRC_DIR) \
		--plugin=protoc-gen-go=$(GOPATH_BIN)/protoc-gen-go \
		--go_out=$(OUT_DIR) \
		--go_opt=paths=source_relative \
		$(M_ARGS) \
		$(PROTO_FILES)

clean:
	rm -rf $(OUT_DIR)
