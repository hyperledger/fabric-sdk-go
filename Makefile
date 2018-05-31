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
# channel-config-[codelevel]-gen: generates the channel configuration transactions and blocks used by tests
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

# Fabric versions used in the Makefile
FABRIC_STABLE_VERSION           := 1.1.0
FABRIC_STABLE_VERSION_MINOR     := 1.1
FABRIC_STABLE_VERSION_MAJOR     := 1
FABRIC_BASEIMAGE_STABLE_VERSION := 0.4.6

FABRIC_PRERELEASE_VERSION       := 1.1.0
FABRIC_PREV_VERSION             := 1.0.6
FABRIC_DEVSTABLE_VERSION_MINOR  := 1.2
FABRIC_DEVSTABLE_VERSION_MAJOR  := 1
FABRIC_PREV_VERSION_MINOR       := 1.0

# Build flags (overridable)
GO_LDFLAGS                 ?=
GO_TESTFLAGS               ?=
FABRIC_SDK_EXPERIMENTAL    ?= true
FABRIC_SDK_EXTRA_GO_TAGS   ?=
FABRIC_SDK_POPULATE_VENDOR ?= true

# Fabric tool versions (overridable)
FABRIC_TOOLS_VERSION ?= $(FABRIC_STABLE_VERSION)
FABRIC_BASE_VERSION  ?= $(FABRIC_BASEIMAGE_STABLE_VERSION)

# Fabric base docker image (overridable)
FABRIC_BASE_IMAGE   ?= hyperledger/fabric-baseimage
FABRIC_BASE_TAG     ?= $(ARCH)-$(FABRIC_BASE_VERSION)

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE ?= hyperledger/fabric-tools
FABRIC_TOOLS_TAG   ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric docker registries (overridable)
FABRIC_RELEASE_REGISTRY     ?= registry.hub.docker.com
FABRIC_DEV_REGISTRY         ?= nexus3.hyperledger.org:10001
FABRIC_DEV_REGISTRY_PRE_CMD ?= docker login -u docker -p docker nexus3.hyperledger.org:10001

# Upstream fabric patching (overridable)
THIRDPARTY_FABRIC_CA_BRANCH ?= master
THIRDPARTY_FABRIC_CA_COMMIT ?= 7c3fc1addc046055f66d45d35a1c47c98364c627
THIRDPARTY_FABRIC_BRANCH    ?= master
THIRDPARTY_FABRIC_COMMIT    ?= 8f79ea1aebdaee1c844d1b5f2c8f89dff18bcffc

# Force removal of images in cleanup (overridable)
FIXTURE_DOCKER_REMOVE_FORCE ?= false

# Options for exercising unit tests (overridable)
FABRIC_SDK_DEPRECATED_UNITTEST   ?= false

# Code levels to exercise integration/e2e tests against (overridable)
FABRIC_STABLE_INTTEST         ?= true
FABRIC_STABLE_PKCS11_INTTEST  ?= false
FABRIC_STABLE_REVOKED_INTTEST ?= false
FABRIC_STABLE_EXPIRED_INTTEST ?= false
FABRIC_PREV_INTTEST           ?= false
FABRIC_PRERELEASE_INTTEST     ?= false
FABRIC_DEVSTABLE_INTTEST      ?= false

# Code levels
FABRIC_STABLE_CODELEVEL_TAG     := stable
FABRIC_PREV_CODELEVEL_TAG       := prev
FABRIC_PRERELEASE_CODELEVEL_TAG := prerelease
FABRIC_DEVSTABLE_CODELEVEL_TAG  := devstable
FABRIC_CODELEVEL_TAG            ?= $(FABRIC_STABLE_CODELEVEL_TAG)

# Code level version targets
FABRIC_STABLE_CODELEVEL_VER     := v$(FABRIC_STABLE_VERSION_MINOR)
FABRIC_PREV_CODELEVEL_VER       := v$(FABRIC_PREV_VERSION_MINOR)
FABRIC_PRERELEASE_CODELEVEL_VER := v$(FABRIC_PRERELEASE_VERSION)
FABRIC_DEVSTABLE_CODELEVEL_VER  := v$(FABRIC_DEVSTABLE_VERSION_MINOR)
FABRIC_CODELEVEL_VER            ?= $(FABRIC_STABLE_CODELEVEL_VER)
FABRIC_CRYPTOCONFIG_VER         ?= v$(FABRIC_STABLE_VERSION_MAJOR)

# Code level to exercise during unit tests
FABRIC_CODELEVEL_UNITTEST_TAG ?= $(FABRIC_DEVSTABLE_CODELEVEL_TAG)
FABRIC_CODELEVEL_UNITTEST_VER ?= $(FABRIC_DEVSTABLE_CODELEVEL_VER)

# Local variables used by makefile
PACKAGE_NAME           := github.com/hyperledger/fabric-sdk-go
ARCH                   := $(shell uname -m)
FIXTURE_PROJECT_NAME   := fabsdkgo
MAKEFILE_THIS          := $(lastword $(MAKEFILE_LIST))
THIS_PATH              := $(patsubst %/,%,$(dir $(abspath $(MAKEFILE_THIS))))
TEST_SCRIPTS_PATH      := test/scripts

# Test fixture paths
FIXTURE_SCRIPTS_PATH   := $(THIS_PATH)/test/scripts
FIXTURE_DOCKERENV_PATH := $(THIS_PATH)/test/fixtures/dockerenv
FIXTURE_SOFTHSM2_PATH  := $(THIS_PATH)/test/fixtures/softhsm2

ifneq ($(GO_LDFLAGS),)
GO_LDFLAGS_ARG := -ldflags=$(GO_LDFLAGS)
else
GO_LDFLAGS_ARG :=
endif

# Fabric tool docker tags at code levels
FABRIC_TOOLS_STABLE_TAG     := $(ARCH)-$(FABRIC_STABLE_VERSION)
FABRIC_TOOLS_PREV_TAG       := $(ARCH)-$(FABRIC_PREV_VERSION)
FABRIC_TOOLS_PRERELEASE_TAG := $(ARCH)-$(FABRIC_PRERELEASE_VERSION)
FABRIC_TOOLS_DEVSTABLE_TAG  := stable

# The version of dep that will be installed by depend-install (or in the CI)
GO_DEP_COMMIT := v0.4.1

# Detect CI
# TODO introduce nightly and adjust verify
ifdef JENKINS_URL
export FABRIC_SDKGO_DEPEND_INSTALL=true

FABRIC_SDK_DEPRECATED_UNITTEST   := false
FABRIC_STABLE_INTTEST            := true
FABRIC_STABLE_PKCS11_INTTEST     := true
FABRIC_STABLE_REVOKED_INTTEST    := true
FABRIC_STABLE_EXPIRED_INTTEST    := true
FABRIC_PREV_INTTEST              := true
FABRIC_PRERELEASE_INTTEST        := false
FABRIC_DEVSTABLE_INTTEST         := true
endif

# Setup Go Tags
GO_TAGS := $(FABRIC_SDK_EXTRA_GO_TAGS)
ifeq ($(FABRIC_SDK_EXPERIMENTAL),true)
GO_TAGS += experimental
endif

# Detect subtarget execution
ifdef FABRIC_SDKGO_SUBTARGET
export FABRIC_SDKGO_DEPEND_INSTALL=false
FABRIC_SDK_POPULATE_VENDOR := false
endif

# DEVSTABLE images are currently only x86_64
ifneq ($(ARCH),x86_64)
FABRIC_DEVSTABLE_INTTEST := false
endif

# Global environment exported for scripts
export GO_CMD
export GO_DEP_CMD
export ARCH
export BASE_ARCH=$(ARCH)
export GO_LDFLAGS
export GO_DEP_COMMIT
export GO_MOCKGEN_COMMIT
export GO_TAGS
export GO_TESTFLAGS
export DOCKER_CMD
export DOCKER_COMPOSE_CMD

.PHONY: all
all: checks unit-test integration-test

.PHONY: depend
depend:
	@$(TEST_SCRIPTS_PATH)/dependencies.sh

.PHONY: depend-install
depend-install:
	@FABRIC_SDKGO_DEPEND_INSTALL="true" $(TEST_SCRIPTS_PATH)/dependencies.sh

.PHONY: checks
checks: depend license lint

.PHONY: license
license:
	@$(TEST_SCRIPTS_PATH)/check_license.sh

.PHONY: lint
lint: populate
	@$(TEST_SCRIPTS_PATH)/check_lint.sh


.PHONY: build-softhsm2-image
build-softhsm2-image:
	 @$(DOCKER_CMD) build --no-cache -q -t "softhsm2-image" \
		--build-arg FABRIC_BASE_IMAGE=$(FABRIC_BASE_IMAGE) \
		--build-arg FABRIC_BASE_TAG=$(FABRIC_BASE_TAG) \
		-f $(FIXTURE_SOFTHSM2_PATH)/Dockerfile .

.PHONY: unit-test
unit-test: checks depend populate
	@FABRIC_SDKGO_CODELEVEL=$(FABRIC_CODELEVEL_UNITTEST_TAG) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_UNITTEST_VER) $(TEST_SCRIPTS_PATH)/unit.sh
ifeq ($(FABRIC_SDK_DEPRECATED_UNITTEST),true)
	@GO_TAGS="$(GO_TAGS) deprecated" GO_TESTFLAGS="$(GO_TESTFLAGS) -count=1" FABRIC_SDKGO_CODELEVEL=$(FABRIC_CODELEVEL_UNITTEST_TAG) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_UNITTEST_VER) $(TEST_SCRIPTS_PATH)/unit.sh
endif

.PHONY: unit-tests
unit-tests: unit-test

.PHONY: unit-tests-pkcs11
unit-tests-pkcs11: checks depend populate
	@FABRIC_SDKGO_CODELEVEL=$(FABRIC_CODELEVEL_UNITTEST_TAG) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_UNITTEST_VER) $(TEST_SCRIPTS_PATH)/unit-pkcs11.sh


.PHONY: integration-tests-stable
integration-tests-stable: clean depend populate
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-prev
integration-tests-prev: clean depend populate
	@. $(FIXTURE_DOCKERENV_PATH)/prev-env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PREV_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PREV_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-prerelease
integration-tests-prerelease: clean depend populate
	@. $(FIXTURE_DOCKERENV_PATH)/prerelease-env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PRERELEASE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PRERELEASE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-devstable
integration-tests-devstable: clean depend populate
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_FIXTURE_VERSION=v$(FABRIC_DEVSTABLE_VERSION_MINOR) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-stable-revoked
integration-tests-stable-revoked: clean depend populate
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-revoked.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-revoked.yaml"

.PHONY: integration-tests-stable-orderer-cert-expired
integration-tests-stable-orderer-cert-expired: clean depend populate
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-expired-orderer.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-expired-orderer.yaml"

.PHONY: integration-tests-stable-peer-cert-expired
integration-tests-stable-peer-cert-expired: clean depend populate
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-expired-peer.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-expired-peer.yaml"



.PHONY: integration-tests-stable-pkcs11
integration-tests-stable-pkcs11: clean depend populate build-softhsm2-image
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-pkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-pkcs11-test.yaml"

# Additional test cases that aren't currently run by the CI
.PHONY: integration-tests-devstable-nomutualtls
integration-tests-devstable-nomutualtls: clean depend populate
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_DOCKERENV_PATH)/nomutualtls-env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ $(FIXTURE_SCRIPTS_PATH)/check_status.sh "-f ./docker-compose.yaml -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests
integration-tests: integration-test

.PHONY: integration-test
integration-test: clean depend populate
ifeq ($(FABRIC_STABLE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable
endif
ifeq ($(FABRIC_STABLE_PKCS11_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-pkcs11
endif
ifeq ($(FABRIC_STABLE_REVOKED_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-revoked
endif

ifeq ($(FABRIC_STABLE_EXPIRED_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-orderer-cert-expired
endif
ifeq ($(FABRIC_STABLE_EXPIRED_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-peer-cert-expired
endif

ifeq ($(FABRIC_PRERELEASE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-prerelease
endif
ifeq ($(FABRIC_DEVSTABLE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-devstable
endif
ifeq ($(FABRIC_PREV_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-prev
endif
	@$(MAKE) -f $(MAKEFILE_THIS) clean

.PHONY: integration-tests-local
integration-tests-local: temp-clean depend populate
	FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_CODELEVEL_TAG) TEST_LOCAL=true  $(TEST_SCRIPTS_PATH)/integration.sh

.PHONY: dockerenv-prev-up
dockerenv-prev-up: clean
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PREV_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PREV_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml up --force-recreate

.PHONY: dockerenv-stable-up
dockerenv-stable-up: clean
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml up --force-recreate

.PHONY: dockerenv-prerelease-up
dockerenv-prerelease-up: clean
	@cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PRERELEASE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PRERELEASE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml up --force-recreate

.PHONY: dockerenv-devstable-up
dockerenv-devstable-up: clean
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY)/ $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml up --force-recreate

.PHONY: dockerenv-latest-up
dockerenv-latest-up: clean
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_DOCKERENV_PATH)/latest-env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY="" $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml up --force-recreate

.PHONY: mock-gen
mock-gen:
	mockgen -build_flags '$(GO_LDFLAGS_ARG)' -package mockcore github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core CryptoSuiteConfig,ConfigBackend,Providers | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockcore/mockcore.gen.go
	mockgen -build_flags '$(GO_LDFLAGS_ARG)' -package mockmsp github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp IdentityConfig,IdentityManager,Providers | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockmsp/mockmsp.gen.go
	mockgen -build_flags '$(GO_LDFLAGS_ARG)' -package mockfab github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab EndpointConfig,ProposalProcessor,Providers | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockfab/mockfab.gen.go
	mockgen -build_flags '$(GO_LDFLAGS_ARG)' -package mockcontext github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context Providers,Client | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockcontext/mockcontext.gen.go
	mockgen -build_flags '$(GO_LDFLAGS_ARG)' -package mocksdkapi github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api CoreProviderFactory,MSPProviderFactory,ServiceProviderFactory | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/fabsdk/test/mocksdkapi/mocksdkapi.gen.go
	mockgen -build_flags '$(GO_LDFLAGS_ARG)' -package mockmspapi github.com/hyperledger/fabric-sdk-go/pkg/msp/api CAClient | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/msp/test/mockmspapi/mockmspapi.gen.go

# TODO - Add cryptogen
.PHONY: channel-config-gen
channel-config-gen:
	@echo "Generating test channel configuration transactions and blocks ..."
	@$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-all-gen
channel-config-all-gen: channel-config-stable-gen channel-config-prev-gen channel-config-prerelease-gen channel-config-devstable-gen

.PHONY: channel-config-stable-gen
channel-config-stable-gen:
	@echo "Generating test channel configuration transactions and blocks (code level stable) ..."
	@$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_STABLE_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_STABLE_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-prev-gen
channel-config-prev-gen:
	@echo "Generating test channel configuration transactions and blocks (code level prev) ..."
	$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_PREV_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_PREV_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-prerelease-gen
channel-config-prerelease-gen:
	@echo "Generating test channel configuration transactions and blocks (code level prerelease) ..."
	$(DOCKER_CMD) run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_PRERELEASE_TAG) \
		/bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_PRERELEASE_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-devstable-gen
channel-config-devstable-gen:
	@echo "Generating test channel configuration transactions and blocks (code level devstable) ..."
	@$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		$(DOCKER_CMD) run -i \
			-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
			$(FABRIC_DEV_REGISTRY)/$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_DEVSTABLE_TAG) \
			/bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_DEVSTABLE_CODELEVEL_VER)/ /opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: thirdparty-pin
thirdparty-pin:
	@echo "Pinning third party packages ..."
	@UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_BRANCH) scripts/third_party_pins/fabric/apply_upstream.sh
	@UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_CA_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_CA_BRANCH) scripts/third_party_pins/fabric-ca/apply_upstream.sh

.PHONY: populate
populate: populate-vendor

.PHONY: populate-vendor
populate-vendor:
ifeq ($(FABRIC_SDK_POPULATE_VENDOR),true)
	@echo "Populating vendor ..."
	@$(GO_DEP_CMD) ensure -vendor-only
endif

.PHONY: populate-clean
populate-clean:
	rm -Rf vendor

.PHONY: temp-clean
temp-clean:
	-rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore /tmp/hfc-kvs /tmp/state /tmp/state-store
	-rm -f integration-report.xml report.xml

.PHONY: clean
clean: temp-clean
	-$(GO_CMD) clean
	-FIXTURE_PROJECT_NAME=$(FIXTURE_PROJECT_NAME) DOCKER_REMOVE_FORCE=$(FIXTURE_DOCKER_REMOVE_FORCE) $(TEST_SCRIPTS_PATH)/clean_integration.sh
