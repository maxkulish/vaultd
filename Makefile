ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BIN_DIR = $(ROOT_DIR)/bin
PROJ_NAME = vaultd
MAC_DIR = $(BIN_DIR)/macOS
LINUX_DIR = $(BIN_DIR)/linux

help: __help__

__help__:
	@echo make build - build go executables in the ./bin folder
	@echo make clean - delete executables, download project from github and build

build: clean
	make build_mac
	make build_linux

build_mac:
	cd $(ROOT_DIR)
	GOOS=darwin GOARCH=amd64 go build --race -o $(MAC_DIR)/$(PROJ_NAME) ./main.go

build_linux:
	cd $(ROOT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_DIR)/$(PROJ_NAME) ./main.go

clean:
	rm -rf ./bin

archive_mac:
	cd $(ROOT_DIR)
	tar -czvf $(MAC_DIR)/$(PROJ_NAME)_mac.tar.gz $(MAC_DIR)/$(PROJ_NAME)

archive_linux:
	cd $(ROOT_DIR)
	tar -czvf $(LINUX_DIR)/$(PROJ_NAME)_linux.tar.gz $(LINUX_DIR)/$(PROJ_NAME)