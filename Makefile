OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
INSTALL_PATH ?= ~/.terraform.d/plugins/github.com/erikh/terraurl/0.0.1/${OS_ARCH}

install:
	mkdir -p ${INSTALL_PATH}
	go build -o ${INSTALL_PATH}/terraform-provider-terraurl
