#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Supported Targets:
# all : runs unit and integration tests
# depend: checks that test dependencies are installed
# depend-install: installs test dependencies
# unit-test: runs all the unit tests
# integration-test: runs all the integration tests
# checks: runs all check conditions (license, spelling, linting)
# clean: stops docker conatainers used for integration testing
# mock-gen: generate mocks needed for testing (using mockgen)
# channel-config-gen: generates the channel configuration transactions and blocks used by tests
# populate: populates generated files (not included in git) - currently only vendor
# populate-vendor: populate the vendor directory based on the lock
# populate-clean: cleans up populated files (might become part of clean eventually) 
# thirdparty-pin: pulls (and patches) pinned dependencies into the project under internal
#

# Tool commands (overridable)
GO_CMD             ?= go
GO_DEP_CMD         ?= dep
DOCKER_CMD         ?= docker
DOCKER_COMPOSE_CMD ?= docker-compose

# Build flags (overridable)
GO_LDFLAGS                 ?= -ldflags=-s
GO_TESTFLAGS               ?=
FABRIC_SDK_EXPERIMENTAL    ?= true
FABRIC_SDK_EXTRA_GO_TAGS   ?=
FABRIC_SDK_POPULATE_VENDOR ?= true

# Fabric tool versions (overridable)
FABRIC_TOOLS_VERSION ?= 1.0.4
FABRIC_BASE_VERSION  ?= 0.4.2

# Fabric base docker image (overridable)
FABRIC_BASE_IMAGE   ?= hyperledger/fabric-baseimage
FABRIC_BASE_TAG     ?= $(ARCH)-$(FABRIC_BASE_VERSION)

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE  ?= hyperledger/fabric-tools
FABRIC_TOOLS_TAG    ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Upstream fabric patching (overridable)
THIRDPARTY_FABRIC_CA_BRANCH ?= master
THIRDPARTY_FABRIC_CA_COMMIT ?= v1.1.0-preview
THIRDPARTY_FABRIC_BRANCH    ?= master
THIRDPARTY_FABRIC_COMMIT    ?= v1.1.0-preview

# Force removal of images in cleanup
FIXTURE_DOCKER_REMOVE_FORCE ?= false

# Local variables used by makefile
PACKAGE_NAME         := github.com/hyperledger/fabric-sdk-go
ARCH                 := $(shell uname -m)
FIXTURE_PROJECT_NAME := fabsdkgo

# The version of dep that will be installed by depend-install (or in the CI)
GO_DEP_COMMIT := v0.3.1

# Setup Go Tags
GO_TAGS := $(FABRIC_SDK_EXTRA_GO_TAGS)
ifeq ($(FABRIC_SDK_EXPERIMENTAL),true)
GO_TAGS += experimental
endif

# Detect CI
ifdef JENKINS_URL
export FABRIC_SDKGO_DEPEND_INSTALL := true
endif

# Global environment exported for scripts
export GO_CMD
export GO_DEP_CMD
export ARCH
export GO_LDFLAGS
export GO_DEP_COMMIT
export GO_TAGS
export GO_TESTFLAGS
export DOCKER_CMD
export DOCKER_COMPOSE_CMD

all: checks unit-test integration-test

depend:
	@test/scripts/dependencies.sh

depend-install:
	@FABRIC_SDKGO_DEPEND_INSTALL="true" test/scripts/dependencies.sh

checks: depend license lint spelling

.PHONY: license build-softhsm2-image
license:
	@test/scripts/check_license.sh

lint: populate
	@test/scripts/check_lint.sh

spelling:
	@test/scripts/check_spelling.sh

build-softhsm2-image:
	 @$(DOCKER_CMD) build --no-cache -q -t "softhsm2-image" \
		--build-arg FABRIC_BASE_IMAGE=$(FABRIC_BASE_IMAGE) \
		--build-arg FABRIC_BASE_TAG=$(FABRIC_BASE_TAG) \
		./test/fixtures/softhsm2

unit-test: checks depend populate
	@test/scripts/unit.sh

unit-tests: unit-test

integration-tests-nopkcs11: clean depend populate
	@cd ./test/fixtures && $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

integration-tests-pkcs11: clean depend populate build-softhsm2-image
	@cd ./test/fixtures && $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-pkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd test/fixtures && ../scripts/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-pkcs11-test.yaml"

integration-test: integration-tests-nopkcs11 integration-tests-pkcs11

mock-gen:
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apitxn ProposalProcessor | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > api/apitxn/mocks/mockapitxn.gen.go
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apiconfig Config | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > api/apiconfig/mocks/mockconfig.gen.go
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apifabca FabricCAClient | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > api/apifabca/mocks/mockfabriccaclient.gen.go

channel-config-gen:
	@echo "Generating test channel configuration transactions and blocks ..."
	@$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		/bin/bash -c "/opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

thirdparty-pin:
	@echo "Pinning third party packages ..."
	@UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_BRANCH) scripts/third_party_pins/fabric/apply_upstream.sh
	@UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_CA_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_CA_BRANCH) scripts/third_party_pins/fabric-ca/apply_upstream.sh

populate: populate-vendor

populate-vendor:
ifeq ($(FABRIC_SDK_POPULATE_VENDOR),true)
	@echo "Populating vendor ..."
	@$(GO_DEP_CMD) ensure -vendor-only
endif

populate-clean:
	rm -Rf vendor

clean:
	-$(GO_CMD) clean
	-rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore /tmp/hfc-kvs /tmp/state
	-rm -f integration-report.xml report.xml
	-FIXTURE_PROJECT_NAME=$(FIXTURE_PROJECT_NAME) DOCKER_REMOVE_FORCE=$(FIXTURE_DOCKER_REMOVE_FORCE) test/scripts/clean_integration.sh