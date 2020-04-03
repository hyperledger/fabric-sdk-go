#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Supported Targets:
# all : runs unit and integration tests
# depend: installs test dependencies
# unit-test: runs all the unit tests
# integration-test: runs all the integration tests
# checks: runs all check conditions (license, spelling, linting)
# clean: stops docker conatainers used for integration testing
# mock-gen: generate mocks needed for testing (using mockgen)
# channel-config-[codelevel]-gen: generates the channel configuration transactions and blocks used by tests
# populate: populates generated files (not included in git) - currently only vendor
# populate-vendor: populate the vendor directory based on the lock
# clean-populate: cleans up populated files (might become part of clean eventually)
# thirdparty-pin: pulls (and patches) pinned dependencies into the project under internal
#

# Tool commands (overridable)
GO_CMD             ?= go
DOCKER_CMD         ?= docker
DOCKER_COMPOSE_CMD ?= docker-compose

# Fabric versions used in the Makefile
FABRIC_STABLE_VERSION           := 1.4.2
FABRIC_STABLE_VERSION_MINOR     := 1.4
FABRIC_STABLE_VERSION_MAJOR     := 1

FABRIC_PRERELEASE_VERSION       :=
FABRIC_PRERELEASE_VERSION_MINOR :=
FABRIC_PREV_VERSION             := 1.3.0
FABRIC_PREV_VERSION_MINOR       := 1.3
FABRIC_DEVSTABLE_VERSION_MINOR  := 1.4
FABRIC_DEVSTABLE_VERSION_MAJOR  := 1

# Build flags (overridable)
GO_LDFLAGS                 ?=
GO_TESTFLAGS               ?=
GO_TESTFLAGS_UNIT          ?= $(GO_TESTFLAGS)
GO_TESTFLAGS_INTEGRATION   ?= $(GO_TESTFLAGS) -failfast
FABRIC_SDK_EXPERIMENTAL    ?= true
FABRIC_SDK_EXTRA_GO_TAGS   ?=
FABRIC_SDK_CHAINCODED      ?= false
FABRIC_SDKGO_TEST_CHANGED  ?= false
FABRIC_SDKGO_TESTRUN_ID    ?= $(shell date +'%Y%m%d%H%M%S')

# Dev tool versions (overridable)
GOLANGCI_LINT_VER ?= v1.23.8

# Fabric tool versions (overridable)
FABRIC_TOOLS_VERSION ?= $(FABRIC_STABLE_VERSION)

# Fabric tools docker image (overridable)
FABRIC_TOOLS_IMAGE ?= hyperledger/fabric-tools
FABRIC_TOOLS_TAG   ?= $(FABRIC_ARCH)-$(FABRIC_TOOLS_VERSION)

# Fabric docker registries (overridable)
FABRIC_RELEASE_REGISTRY     ?=
FABRIC_DEV_REGISTRY         ?= nexus3.hyperledger.org:10001
FABRIC_DEV_REGISTRY_PRE_CMD ?= docker login -u docker -p docker nexus3.hyperledger.org:10001

# Base image variables for socat and softshm builds
BASE_GO_VERSION = "1.13"

# Upstream fabric patching (overridable)
THIRDPARTY_FABRIC_CA_BRANCH ?= master
THIRDPARTY_FABRIC_CA_COMMIT ?= 02fe02b0a6f224aac8ac6fd813cecc590ec2a024
THIRDPARTY_FABRIC_BRANCH    ?= master
THIRDPARTY_FABRIC_COMMIT    ?= v2.0.0-beta

# Force removal of images in cleanup (overridable)
FIXTURE_DOCKER_REMOVE_FORCE ?= false

# Options for exercising unit tests (overridable)
FABRIC_SDK_DEPRECATED_UNITTEST ?= false

# Code levels to exercise integration/e2e tests against (overridable)
FABRIC_STABLE_INTTEST          ?= true
FABRIC_STABLE_PKCS11_INTTEST   ?= false
FABRIC_STABLE_NEGATIVE_INTTEST ?= false
FABRIC_PREV_INTTEST            ?= false
FABRIC_PRERELEASE_INTTEST      ?= false
FABRIC_DEVSTABLE_INTTEST       ?= false
FABRIC_STABLE_LOCAL_INTTEST    ?= false
FABRIC_DEVSTABLE_LOCAL_INTTEST ?= false

# Code levels
FABRIC_STABLE_CODELEVEL_TAG     := stable
FABRIC_PREV_CODELEVEL_TAG       := prev
FABRIC_PRERELEASE_CODELEVEL_TAG := prerelease
FABRIC_DEVSTABLE_CODELEVEL_TAG  := devstable
FABRIC_CODELEVEL_TAG            ?= $(FABRIC_STABLE_CODELEVEL_TAG)

# Code level version targets
FABRIC_STABLE_CODELEVEL_VER     := v$(FABRIC_STABLE_VERSION_MINOR)
FABRIC_PREV_CODELEVEL_VER       := v$(FABRIC_PREV_VERSION_MINOR)
FABRIC_PRERELEASE_CODELEVEL_VER := v$(FABRIC_PRERELEASE_VERSION_MINOR)
FABRIC_DEVSTABLE_CODELEVEL_VER  := v$(FABRIC_DEVSTABLE_VERSION_MINOR)
FABRIC_CODELEVEL_VER            ?= $(FABRIC_STABLE_CODELEVEL_VER)
FABRIC_CRYPTOCONFIG_VER         ?= v$(FABRIC_STABLE_VERSION_MAJOR)

# Code level to exercise during unit tests
FABRIC_CODELEVEL_UNITTEST_TAG ?= $(FABRIC_STABLE_CODELEVEL_TAG)
FABRIC_CODELEVEL_UNITTEST_VER ?= $(FABRIC_STABLE_CODELEVEL_VER)

# Local variables used by makefile
PROJECT_NAME           := fabric-sdk-go
ARCH                   := $(shell uname -m)
OS_NAME                := $(shell uname -s)
FIXTURE_PROJECT_NAME   := fabsdkgo
MAKEFILE_THIS          := $(lastword $(MAKEFILE_LIST))
THIS_PATH              := $(patsubst %/,%,$(dir $(abspath $(MAKEFILE_THIS))))
TEST_SCRIPTS_PATH      := test/scripts
SOCAT_DOCKER_IMG       := $(shell docker images -q fabsdkgo-socat 2> /dev/null)

# Tool commands
MOCKGEN_CMD := gobin -run github.com/golang/mock/mockgen

# Test fixture paths
FIXTURE_SCRIPTS_PATH      := $(THIS_PATH)/test/scripts
FIXTURE_DOCKERENV_PATH    := $(THIS_PATH)/test/fixtures/dockerenv
FIXTURE_CRYPTOCONFIG_PATH := $(THIS_PATH)/test/fixtures/fabric/$(FABRIC_CRYPTOCONFIG_VER)/crypto-config
FIXTURE_SOFTHSM2_PATH     := $(THIS_PATH)/test/fixtures/softhsm2
FIXTURE_SOCAT_PATH        := $(THIS_PATH)/test/fixtures/socat

ifneq ($(GO_LDFLAGS),)
GO_LDFLAGS_ARG := -ldflags=$(GO_LDFLAGS)
else
GO_LDFLAGS_ARG :=
endif

ifneq ($(FABRIC_RELEASE_REGISTRY),)
FABRIC_RELEASE_REGISTRY := $(FABRIC_RELEASE_REGISTRY)/
endif

ifneq ($(FABRIC_DEV_REGISTRY),)
FABRIC_DEV_REGISTRY := $(FABRIC_DEV_REGISTRY)/
endif

# Fabric tool docker tags at code levels
FABRIC_TOOLS_STABLE_TAG     = $(FABRIC_ARCH)-$(FABRIC_STABLE_VERSION)
FABRIC_TOOLS_PREV_TAG       = $(FABRIC_ARCH)-$(FABRIC_PREV_VERSION)
FABRIC_TOOLS_PRERELEASE_TAG = $(FABRIC_ARCH)-$(FABRIC_PRERELEASE_VERSION)
FABRIC_TOOLS_DEVSTABLE_TAG  := stable

# Detect CI
# TODO introduce nightly and adjust verify
ifdef JENKINS_URL
export FABRIC_SDKGO_DEPEND_INSTALL=true
FABRIC_SDK_CHAINCODED            := true
# TODO: disabled FABRIC_SDKGO_TEST_CHANGED optimization - while tests are being fixed.
FABRIC_SDKGO_TEST_CHANGED        := false
FABRIC_SDK_DEPRECATED_UNITTEST   := false
FABRIC_STABLE_INTTEST            := true
FABRIC_STABLE_PKCS11_INTTEST     := true
FABRIC_STABLE_NEGATIVE_INTTEST   := true
FABRIC_PREV_INTTEST              := true
FABRIC_PRERELEASE_INTTEST        := false
FABRIC_DEVSTABLE_INTTEST         := false
FABRIC_STABLE_LOCAL_INTTEST      := false
FABRIC_DEVSTABLE_LOCAL_INTTEST   := false
endif

# Determine if use mock chaincode daemon should be used
FABRIC_SDKGO_ENABLE_CHAINCODED := false
#chaincoded is currently able to intercept the docker calls without need for forwarding.
#(so reverse proxy to docker via socat is currently disabled).
#ifneq ($(SOCAT_DOCKER_IMG),)
ifeq ($(FABRIC_SDK_CHAINCODED),true)
FABRIC_SDKGO_ENABLE_CHAINCODED := true
#endif
endif

# Determine if internal dependency calc should be used
# If so, disable GOCACHE
ifeq ($(FABRIC_SDKGO_TEST_CHANGED),true)
ifeq (,$(findstring $(GO_TESTFLAGS_UNIT),-count=1))
GO_TESTFLAGS_UNIT += -count=1
endif
ifeq (,$(findstring $(GO_TESTFLAGS_INTEGRATION),-count=1))
GO_TESTFLAGS_INTEGRATION += -count=1
endif
endif

# Setup Go Tags
GO_TAGS := $(FABRIC_SDK_EXTRA_GO_TAGS)
ifeq ($(FABRIC_SDK_EXPERIMENTAL),true)
GO_TAGS += experimental
endif

# Detect subtarget execution
ifdef FABRIC_SDKGO_SUBTARGET
export FABRIC_SDKGO_DEPEND_INSTALL=false
endif

FABRIC_ARCH := $(ARCH)

ifneq ($(ARCH),x86_64)
# DEVSTABLE images are currently only x86_64
FABRIC_DEVSTABLE_INTTEST := false
else
# Recent Fabric builds follow GOARCH (e.g., amd64)
FABRIC_ARCH := amd64
endif

# Docker-compose
BASE_DOCKER_COMPOSE_FILES := -f ./docker-compose.yaml
ifeq ($(FABRIC_SDKGO_ENABLE_CHAINCODED),true)
BASE_DOCKER_COMPOSE_FILES := -f ./docker-compose-chaincoded.yaml $(BASE_DOCKER_COMPOSE_FILES)
export CORE_VM_ENDPOINT=http://chaincoded.example.com:9375
else
BASE_DOCKER_COMPOSE_FILES := -f ./docker-compose-std.yaml $(BASE_DOCKER_COMPOSE_FILES)
endif
DOCKER_COMPOSE_UP_FLAGS            := --remove-orphans --force-recreate
DOCKER_COMPOSE_UP_TEST_FLAGS       := $(DOCKER_COMPOSE_UP_FLAGS) --abort-on-container-exit
DOCKER_COMPOSE_UP_BACKGROUND_FLAGS := $(DOCKER_COMPOSE_UP_FLAGS) -d
DOCKER_COMPOSE_UP_STANDALONE_FLAGS := $(DOCKER_COMPOSE_UP_FLAGS)
DOCKER_COMPOSE_PULL_FLAGS :=

# Global environment exported for scripts
export GO_CMD
export ARCH
export FABRIC_ARCH
export GO_LDFLAGS
export GO_MOCKGEN_COMMIT
export GO_TAGS
export DOCKER_CMD
export DOCKER_COMPOSE_CMD
export FABRIC_SDKGO_TESTRUN_ID
export GO111MODULE=on

.PHONY: all
all: version depend-noforce license unit-test integration-test

.PHONY: version
version:
	@$(TEST_SCRIPTS_PATH)/check_version.sh

.PHONY: depend
depend: version
	@$(TEST_SCRIPTS_PATH)/dependencies.sh -f

.PHONY: depend-noforce
depend-noforce: version
ifeq ($(FABRIC_SDKGO_DEPEND_INSTALL),true)
	@$(TEST_SCRIPTS_PATH)/dependencies.sh
	@$(TEST_SCRIPTS_PATH)/dependencies.sh -c
else
	-@$(TEST_SCRIPTS_PATH)/dependencies.sh -c
endif

.PHONY: checks
checks: version depend-noforce license lint

.PHONY: license
license: version
	@$(TEST_SCRIPTS_PATH)/check_license.sh

.PHONY: lint
lint: version populate-noforce lint-submodules
	@MODULE="github.com/hyperledger/fabric-sdk-go" PKG_ROOT="./pkg" LINT_CHANGED_ONLY=true GOLANGCI_LINT_VER=$(GOLANGCI_LINT_VER) $(TEST_SCRIPTS_PATH)/check_lint.sh

.PHONY: lint-submodules
lint-submodules: version populate-noforce
	@MODULE="github.com/hyperledger/fabric-sdk-go/test/integration" LINT_CHANGED_ONLY=true GOLANGCI_LINT_VER=$(GOLANGCI_LINT_VER) $(TEST_SCRIPTS_PATH)/check_lint.sh
	@MODULE="github.com/hyperledger/fabric-sdk-go/test/performance" LINT_CHANGED_ONLY=true GOLANGCI_LINT_VER=$(GOLANGCI_LINT_VER) $(TEST_SCRIPTS_PATH)/check_lint.sh

.PHONY: lint-all
lint-all: version populate-noforce
	@MODULE="github.com/hyperledger/fabric-sdk-go" PKG_ROOT="./pkg" GOLANGCI_LINT_VER=$(GOLANGCI_LINT_VER) $(TEST_SCRIPTS_PATH)/check_lint.sh
	@MODULE="github.com/hyperledger/fabric-sdk-go/test/integration" GOLANGCI_LINT_VER=$(GOLANGCI_LINT_VER) $(TEST_SCRIPTS_PATH)/check_lint.sh
	@MODULE="github.com/hyperledger/fabric-sdk-go/test/performance" GOLANGCI_LINT_VER=$(GOLANGCI_LINT_VER) $(TEST_SCRIPTS_PATH)/check_lint.sh

.PHONY: build-softhsm2-image
build-softhsm2-image:
	 @$(DOCKER_CMD) build --no-cache -q -t "fabsdkgo-softhsm2" \
		--build-arg BASE_GO_VERSION=$(BASE_GO_VERSION) \
		-f $(FIXTURE_SOFTHSM2_PATH)/Dockerfile .

.PHONY: unit-test
unit-test: clean-tests depend-noforce populate-noforce license lint-submodules
	@TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) TEST_WITH_LINTER=true FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_CODELEVEL_UNITTEST_TAG) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_UNITTEST_VER) \
	GO_TESTFLAGS="$(GO_TESTFLAGS_UNIT)" \
	GOLANGCI_LINT_VER="$(GOLANGCI_LINT_VER)" \
	MODULE="github.com/hyperledger/fabric-sdk-go" \
	PKG_ROOT="./pkg" \
	$(TEST_SCRIPTS_PATH)/unit.sh
ifeq ($(FABRIC_SDK_DEPRECATED_UNITTEST),true)
	@GO_TAGS="$(GO_TAGS) deprecated" TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_CODELEVEL_UNITTEST_TAG) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_UNITTEST_VER) \
	GOLANGCI_LINT_VER="$(GOLANGCI_LINT_VER)" \
	MODULE="github.com/hyperledger/fabric-sdk-go" \
	PKG_ROOT="./pkg" \
	$(TEST_SCRIPTS_PATH)/unit.sh
endif

.PHONY: unit-tests
unit-tests: unit-test

.PHONY: unit-tests-pkcs11
unit-tests-pkcs11: clean-tests depend-noforce populate-noforce license
	@TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) TEST_WITH_LINTER=true FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_CODELEVEL_UNITTEST_TAG) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_UNITTEST_VER) \
	GO_TESTFLAGS="$(GO_TESTFLAGS_UNIT)" \
	GOLANGCI_LINT_VER="$(GOLANGCI_LINT_VER)" \
	MODULE="github.com/hyperledger/fabric-sdk-go" \
	PKG_ROOT="./pkg" \
	$(TEST_SCRIPTS_PATH)/unit-pkcs11.sh

.PHONY: integration-tests-stable
integration-tests-stable: clean-tests depend-noforce populate-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/stable-env.sh && \
	    . $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
	    cd $(FIXTURE_DOCKERENV_PATH) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-nopkcs11-test.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-prev
integration-tests-prev: clean-tests depend-noforce populate-noforce populate-fixtures-prev-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/prev-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) E2E_ONLY="false" FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PREV_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PREV_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-nopkcs11-test.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-prerelease
integration-tests-prerelease: clean-tests depend-noforce populate-noforce populate-fixtures-prerelease-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/prerelease-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PRERELEASE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PRERELEASE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-nopkcs11-test.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-devstable
integration-tests-devstable: clean-tests depend-noforce populate-noforce populate-fixtures-devstable-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) $(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) pull $(DOCKER_COMPOSE_PULL_FLAGS) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_FIXTURE_VERSION=v$(FABRIC_DEVSTABLE_VERSION_MINOR) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-nopkcs11-test.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests-stable-negative
integration-tests-stable-negative: clean-tests depend-noforce populate-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/stable-env.sh && \
	    . $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-negative.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-negative.yaml"

.PHONY: integration-tests-stable-pkcs11
integration-tests-stable-pkcs11: clean-tests depend-noforce populate-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/stable-env.sh && \
	    . $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-pkcs11-test.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-pkcs11-test.yaml"

# Additional test cases that aren't currently run by the CI
.PHONY: integration-tests-devstable-nomutualtls
integration-tests-devstable-nomutualtls: clean-tests depend-noforce populate-noforce populate-fixtures-devstable-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_DOCKERENV_PATH)/nomutualtls-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) $(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) pull $(DOCKER_COMPOSE_PULL_FLAGS) && \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) -f docker-compose-nopkcs11-test.yaml up $(DOCKER_COMPOSE_UP_TEST_FLAGS)
	@cd $(FIXTURE_DOCKERENV_PATH) && FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) $(FIXTURE_SCRIPTS_PATH)/check_status.sh "$(BASE_DOCKER_COMPOSE_FILES) -f ./docker-compose-nopkcs11-test.yaml"

.PHONY: integration-tests
integration-tests: integration-test

.PHONY: integration-test
integration-test: clean-tests depend-noforce populate-noforce
ifeq ($(FABRIC_STABLE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable
endif
ifeq ($(FABRIC_STABLE_PKCS11_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-pkcs11
endif
ifeq ($(FABRIC_STABLE_NEGATIVE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-negative
endif

ifeq ($(FABRIC_PRERELEASE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-prerelease
endif
ifeq ($(FABRIC_DEVSTABLE_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-devstable
endif
ifeq ($(FABRIC_PREV_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-prev
endif
ifeq ($(FABRIC_STABLE_LOCAL_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-stable-local
endif
ifeq ($(FABRIC_DEVSTABLE_LOCAL_INTTEST),true)
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests
	@FABRIC_SDKGO_SUBTARGET=true $(MAKE) -f $(MAKEFILE_THIS) integration-tests-devstable-local
endif
	@$(MAKE) -f $(MAKEFILE_THIS) clean-tests

.PHONY: integration-tests-local
integration-tests-local: clean-tests-temp depend-noforce populate-noforce
	TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_CODELEVEL_TAG) TEST_LOCAL=true  $(TEST_SCRIPTS_PATH)/integration.sh

.PHONY: integration-tests-stable-local
integration-tests-stable-local: clean-tests-temp depend-noforce populate-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/stable-env.sh && \
	    . $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_BACKGROUND_FLAGS)
	FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_CODELEVEL_TAG) TEST_LOCAL=true  $(TEST_SCRIPTS_PATH)/integration.sh
	@cd $(FIXTURE_DOCKERENV_PATH) && $(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) down

.PHONY: integration-tests-devstable-local
integration-tests-devstable-local: clean-tests-temp depend-noforce populate-noforce populate-fixtures-devstable-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) $(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) pull $(DOCKER_COMPOSE_PULL_FLAGS) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) \
		TEST_CHANGED_ONLY=$(FABRIC_SDKGO_TEST_CHANGED) GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_BACKGROUND_FLAGS)
	FABRIC_FIXTURE_VERSION=v$(FABRIC_DEVSTABLE_VERSION_MINOR) FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) TEST_LOCAL=true  $(TEST_SCRIPTS_PATH)/integration.sh
	@cd $(FIXTURE_DOCKERENV_PATH) && $(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) down

.PHONY: dockerenv-prev-up
dockerenv-prev-up: clean-tests populate-fixtures-prev-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/prev-env.sh && \
		$(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PREV_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PREV_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		GO_TESTFLAGS="$(GO_TESTFLAGS_INTEGRATION)" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_STANDALONE_FLAGS)

.PHONY: dockerenv-stable-up
dockerenv-stable-up: clean-tests populate-fixtures-stable-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/stable-env.sh && \
	    . $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_STABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_STANDALONE_FLAGS)

.PHONY: dockerenv-prerelease-up
dockerenv-prerelease-up: clean-tests populate-fixtures-prerelease-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/prerelease-env.sh && \
		$(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_PRERELEASE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PRERELEASE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_RELEASE_REGISTRY) \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_STANDALONE_FLAGS)

.PHONY: dockerenv-devstable-up
dockerenv-devstable-up: clean-tests populate-fixtures-devstable-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) $(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) pull $(DOCKER_COMPOSE_PULL_FLAGS) && \
 		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY=$(FABRIC_DEV_REGISTRY) \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_STANDALONE_FLAGS)

.PHONY: dockerenv-latest-up
dockerenv-latest-up: clean-tests populate-fixtures-devstable-noforce
	@. $(FIXTURE_DOCKERENV_PATH)/devstable-env.sh && \
		. $(FIXTURE_DOCKERENV_PATH)/latest-env.sh && \
		. $(FIXTURE_CRYPTOCONFIG_PATH)/env.sh && \
		cd $(FIXTURE_DOCKERENV_PATH) && \
		FABRIC_SDKGO_CODELEVEL_VER=$(FABRIC_DEVSTABLE_CODELEVEL_VER) FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) FABRIC_DOCKER_REGISTRY="" \
		$(DOCKER_COMPOSE_CMD) $(BASE_DOCKER_COMPOSE_FILES) up $(DOCKER_COMPOSE_UP_STANDALONE_FLAGS)

.PHONY: mock-gen
mock-gen:
	$(MOCKGEN_CMD) -build_flags '$(GO_LDFLAGS_ARG)' -package mockcore github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core CryptoSuiteConfig,ConfigBackend,Providers | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockcore/mockcore.gen.go
	$(MOCKGEN_CMD) -build_flags '$(GO_LDFLAGS_ARG)' -package mockmsp github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp IdentityConfig,IdentityManager,Providers | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockmsp/mockmsp.gen.go
	$(MOCKGEN_CMD) -build_flags '$(GO_LDFLAGS_ARG)' -package mockfab github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab EndpointConfig,ProposalProcessor,Providers | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockfab/mockfab.gen.go
	$(MOCKGEN_CMD) -build_flags '$(GO_LDFLAGS_ARG)' -package mockcontext github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context Providers,Client | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/common/providers/test/mockcontext/mockcontext.gen.go
	$(MOCKGEN_CMD) -build_flags '$(GO_LDFLAGS_ARG)' -package mocksdkapi github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api CoreProviderFactory,MSPProviderFactory,ServiceProviderFactory | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/fabsdk/test/mocksdkapi/mocksdkapi.gen.go
	$(MOCKGEN_CMD) -build_flags '$(GO_LDFLAGS_ARG)' -package mockmspapi github.com/hyperledger/fabric-sdk-go/pkg/msp/api CAClient | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g" | goimports > pkg/msp/test/mockmspapi/mockmspapi.gen.go

.PHONY: crypto-gen
crypto-gen:
	@echo "Generating crypto directory ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_CRYPTOCONFIG_VER) /opt/workspace/${PROJECT_NAME}/test/scripts/generate_crypto.sh"

.PHONY: channel-config-gen
channel-config-gen:
	@echo "Generating test channel configuration transactions and blocks ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_CODELEVEL_VER)/ /opt/workspace/${PROJECT_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-all-gen
channel-config-all-gen: channel-config-stable-gen channel-config-prev-gen channel-config-prerelease-gen channel-config-devstable-gen

.PHONY: channel-config-stable-gen
channel-config-stable-gen:
	@echo "Generating test channel configuration transactions and blocks (code level stable) ..."
	@$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_STABLE_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_STABLE_CODELEVEL_VER)/ /opt/workspace/${PROJECT_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-prev-gen
channel-config-prev-gen:
	@echo "Generating test channel configuration transactions and blocks (code level prev) ..."
	$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_PREV_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_PREV_CODELEVEL_VER)/ /opt/workspace/${PROJECT_NAME}/test/scripts/generate_channeltx.sh"

.PHONY: channel-config-prerelease-gen
channel-config-prerelease-gen:
ifneq ($(FABRIC_PRERELEASE_VERSION),)
	@echo "Generating test channel configuration transactions and blocks (code level prerelease) ..."
	$(DOCKER_CMD) run -i \
		-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
		$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_PRERELEASE_TAG) \
		//bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_PRERELEASE_CODELEVEL_VER)/ /opt/workspace/${PROJECT_NAME}/test/scripts/generate_channeltx.sh"
endif

.PHONY: channel-config-devstable-gen
channel-config-devstable-gen:
ifeq ($(ARCH),x86_64)
	@echo "Generating test channel configuration transactions and blocks (code level devstable) ..."
	@$(FABRIC_DEV_REGISTRY_PRE_CMD) && \
		$(DOCKER_CMD) run -i \
			-v /$(abspath .):/opt/workspace/$(PROJECT_NAME) -u $(shell id -u):$(shell id -g) \
			$(FABRIC_DEV_REGISTRY)$(FABRIC_TOOLS_IMAGE):$(FABRIC_TOOLS_DEVSTABLE_TAG) \
			//bin/bash -c "FABRIC_VERSION_DIR=fabric/$(FABRIC_DEVSTABLE_CODELEVEL_VER)/ /opt/workspace/${PROJECT_NAME}/test/scripts/generate_channeltx.sh"
endif

.PHONY: thirdparty-pin
thirdparty-pin:
	@echo "Pinning third party packages ..."
	@THIRDPARTY_FABRIC_COMMIT=$(THIRDPARTY_FABRIC_COMMIT) \
	THIRDPARTY_FABRIC_BRANCH=$(THIRDPARTY_FABRIC_BRANCH) \
	THIRDPARTY_FABRIC_CA_COMMIT=$(THIRDPARTY_FABRIC_CA_COMMIT) \
	THIRDPARTY_FABRIC_CA_BRANCH=$(THIRDPARTY_FABRIC_CA_BRANCH) \
	scripts/third_party_pins/apply_thirdparty_pins.sh

.PHONY: populate
populate: populate-vendor populate-fixtures-stable

.PHONY: populate-vendor
populate-vendor:
	@go mod vendor

.PHONY: populate-fixtures-stable
populate-fixtures-stable:
	@FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) \
	FABRIC_FIXTURE_VERSION=v$(FABRIC_STABLE_VERSION_MINOR) \
	FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) \
	$(TEST_SCRIPTS_PATH)/populate-fixtures.sh -f

.PHONY: populate-noforce
populate-noforce: populate-fixtures-stable-noforce

.PHONY: populate-fixtures-stable-noforce
populate-fixtures-stable-noforce:
	@FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) \
	FABRIC_FIXTURE_VERSION=v$(FABRIC_STABLE_VERSION_MINOR) \
	FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_STABLE_CODELEVEL_TAG) \
	$(TEST_SCRIPTS_PATH)/populate-fixtures.sh

.PHONY: populate-fixtures-prev-noforce
populate-fixtures-prev-noforce:
	@FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) \
	FABRIC_FIXTURE_VERSION=v$(FABRIC_PREV_VERSION_MINOR) \
	FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PREV_CODELEVEL_TAG) \
	$(TEST_SCRIPTS_PATH)/populate-fixtures.sh

.PHONY: populate-fixtures-prerelease-noforce
populate-fixtures-prerelease-noforce:
	@FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) \
	FABRIC_FIXTURE_VERSION=v$(FABRIC_PRERELEASE_VERSION_MINOR) \
	FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_PRERELEASE_CODELEVEL_TAG) \
	$(TEST_SCRIPTS_PATH)/populate-fixtures.sh

.PHONY: populate-fixtures-devstable-noforce
populate-fixtures-devstable-noforce:
	@FABRIC_CRYPTOCONFIG_VERSION=$(FABRIC_CRYPTOCONFIG_VER) \
	FABRIC_FIXTURE_VERSION=v$(FABRIC_DEVSTABLE_VERSION_MINOR) \
	FABRIC_SDKGO_CODELEVEL_TAG=$(FABRIC_DEVSTABLE_CODELEVEL_TAG) \
	$(TEST_SCRIPTS_PATH)/populate-fixtures.sh


.PHONY: clean
clean: clean-tests clean-fixtures clean-cache clean-populate

.PHONY: clean-populate
clean-populate:
	rm -Rf vendor
	rm -Rf scripts/_go/src/chaincoded/vendor

.PHONY: clean-cache
clean-cache:
ifeq ($(OS_NAME),Darwin)
	rm -Rf ${HOME}/Library/Caches/fabric-sdk-go
else
	rm -Rf ${HOME}/.cache/fabric-sdk-go
endif

.PHONY: clean-depend-images
clean-depend-images: clean-tests
	docker rmi -f fabsdkgo-socat
	docker rmi -f fabsdkgo-softhsm2

.PHONY: clean-fixtures
clean-fixtures:
	-rm -Rf test/fixtures/fabric/*/crypto-config
	-rm -Rf test/fixtures/fabric/*/channel

.PHONY: clean-tests-build
clean-tests-build:
	-$(GO_CMD) clean
	-FIXTURE_PROJECT_NAME=$(FIXTURE_PROJECT_NAME) DOCKER_REMOVE_FORCE=$(FIXTURE_DOCKER_REMOVE_FORCE) $(TEST_SCRIPTS_PATH)/clean_integration.sh

.PHONY: clean-tests-temp
clean-tests-temp:
	-rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore /tmp/hfc-kvs /tmp/state /tmp/state-store
	-rm -f integration-report.xml report.xml

.PHONY: clean-tests
clean-tests: clean-tests-temp clean-tests-build
