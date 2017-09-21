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
#
# Instructions to generate .tx files used for creating channels:
# Download the configtxgen binary for your OS from (it is located in the .tar.gz file):
# https://nexus.hyperledger.org/content/repositories/releases/org/hyperledger/fabric/hyperledger-fabric
# Sample command: $ path/to/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx testchannel.tx -channelID testchannel
# More Docs: http://hyperledger-fabric.readthedocs.io/en/latest/configtxgen.html
#

# Tool commands
GO_CMD             ?= go
GO_DEP_CMD         ?= dep
DOCKER_CMD         ?= docker
DOCKER_COMPOSE_CMD ?= docker-compose

# Build flags
ARCH         := $(shell uname -m)
GO_LDFLAGS   ?= -ldflags=-s
EXPERIMENTAL ?= true  # includes experimental features in the tests
BRANCHFAB    ?= false # requires testing against fabric with cherry picks from gerrit

ifeq ($(EXPERIMENTAL),true)
GO_TAGS += experimental
endif

ifeq ($(BRANCHFAB),true)
GO_TAGS += branchfab
endif

# Upstream fabric patching
THIRDPARTY_FABRIC_CA_BRANCH ?= release
THIRDPARTY_FABRIC_CA_COMMIT ?= v1.0.2
THIRDPARTY_FABRIC_BRANCH    ?= master
THIRDPARTY_FABRIC_COMMIT    ?= a657db28a0ff53ed512bd6f4ac4786a0f4ca709c

# Tool versions
GO_DEP_COMMIT        := v0.3.0 # the version of dep that will be installed by depend-install (or in the CI)
FABRIC_TOOLS_VERSION ?= 1.0.1
FABRIC_BASE_VERSION  ?= 0.3.1

# Fabric Base Docker Image
FABRIC_BASE_IMAGE   ?= hyperledger/fabric-baseimage
FABRIC_BASE_TAG     ?= $(ARCH)-$(FABRIC_BASE_VERSION)

# Fabric Tools Docker Image
FABRIC_TOOLS_IMAGE  ?= hyperledger/fabric-tools
FABRIC_TOOLS_TAG    ?= $(ARCH)-$(FABRIC_TOOLS_VERSION)

# Local variables used by makefile
PACKAGE_NAME=github.com/hyperledger/fabric-sdk-go

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
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apitxn ProposalProcessor | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g"  > api/apitxn/mocks/mockapitxn.gen.go
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apiconfig Config | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g"  > api/apiconfig/mocks/mockconfig.gen.go
	mockgen -build_flags '$(GO_LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apifabca FabricCAClient | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g"  > api/apifabca/mocks/mockfabriccaclient.gen.go

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
	@echo "Populating vendor ..."
	@$(GO_DEP_CMD) ensure -vendor-only

populate-clean:
	rm -Rf vendor

clean:
	$(GO_CMD) clean
	rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore
	rm -f integration-report.xml report.xml
	rm -f test/fixtures/tls/fabricca/certs/server/ca.org*.example.com-cert.pem
	cd test/fixtures && $(DOCKER_COMPOSE_CMD) -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml -f docker-compose-pkcs11-test.yaml down
