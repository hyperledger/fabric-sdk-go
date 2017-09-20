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
# channel-artifacts: generates the channel configuration transactions and blocks used by tests
# populate: populates generated files (not included in git) - currently only vendor
# populate-vendor: populate the vendor directory based on the lock
# populate-clean: cleans up populated files (might become part of clean eventually)
# thirdparty-pin: pulls (and patches) pinned dependencies into the project under internal
#


export ARCH=$(shell uname -m)
export LDFLAGS=-ldflags=-s
export DOCKER_NS=hyperledger
export DOCKER_TAG=$(ARCH)-0.3.1
export GO_DEP_COMMIT=v0.3.0 # the version of dep that will be installed by depend-install (or in the CI)

# Upstream fabric patching
THIRDPARTY_FABRIC_CA_BRANCH=release
THIRDPARTY_FABRIC_CA_COMMIT=v1.0.1
THIRDPARTY_FABRIC_BRANCH=master
THIRDPARTY_FABRIC_COMMIT=a657db28a0ff53ed512bd6f4ac4786a0f4ca709c

# Local variables used by makefile
PACKAGE_NAME=github.com/hyperledger/fabric-sdk-go
FABRIC_TOOLS_RELEASE=1.0.1

# Detect CI
ifdef JENKINS_URL
export FABRIC_SDKGO_DEPEND_INSTALL=true
endif

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

edit-docker:
	@cd ./test/fixtures && sed -i.bak -e 's/_NS_/$(DOCKER_NS)/g' Dockerfile\
	&& sed -i.bak -e 's/_TAG_/$(DOCKER_TAG)/g'  Dockerfile\
	&& rm -rf Dockerfile.bak

build-softhsm2-image:
	 @cd ./test/fixtures && docker build --no-cache -q  -t "softhsm2-image" . \

restore-docker-file:
	@cd ./test/fixtures && sed -i.bak -e 's/$(DOCKER_NS)/_NS_/g' Dockerfile\
	&& sed -i.bak -e 's/$(DOCKER_TAG)/_TAG_/g'  Dockerfile\
	&& rm -rf Dockerfile.bak

unit-test: checks depend populate
	@test/scripts/unit.sh

unit-tests: unit-test

integration-tests: clean depend populate edit-docker build-softhsm2-image restore-docker-file
	@cd ./test/fixtures && docker-compose -f docker-compose.yaml -f docker-compose-integration-test.yaml up --force-recreate --abort-on-container-exit
	@test/scripts/check_status.sh "-f ./test/fixtures/docker-compose.yaml -f ./test/fixtures/docker-compose-integration-test.yaml"

integration-tests-pkcs11: clean depend populate edit-docker build-softhsm2-image restore-docker-file
	@cd ./test/fixtures && docker-compose -f docker-compose.yaml -f docker-compose-pkcs11-test.yaml up --force-recreate --abort-on-container-exit
	@test/scripts/check_status.sh "-f ./test/fixtures/docker-compose.yaml -f ./test/fixtures/docker-compose-pkcs11-test.yaml"

integration-test: integration-tests integration-tests-pkcs11

mock-gen:
	mockgen -build_flags '$(LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apitxn ProposalProcessor | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g"  > api/apitxn/mocks/mockapitxn.gen.go
	mockgen -build_flags '$(LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apiconfig Config | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g"  > api/apiconfig/mocks/mockconfig.gen.go
	mockgen -build_flags '$(LDFLAGS)' github.com/hyperledger/fabric-sdk-go/api/apifabca FabricCAClient | sed "s/github.com\/hyperledger\/fabric-sdk-go\/vendor\///g"  > api/apifabca/mocks/mockfabriccaclient.gen.go

channel-artifacts:
	@echo "Generating test channel configuration transactions and blocks"
	@docker run -i \
		-v $(abspath .):/opt/gopath/src/$(PACKAGE_NAME) \
		hyperledger/fabric-tools:$(ARCH)-$(FABRIC_TOOLS_RELEASE) \
		/bin/bash -c "/opt/gopath/src/${PACKAGE_NAME}/test/scripts/generate_channeltx.sh"

thirdparty-pin:
	UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_BRANCH) scripts/third_party_pins/fabric/apply_upstream.sh
	UPSTREAM_COMMIT=$(THIRDPARTY_FABRIC_CA_COMMIT) UPSTREAM_BRANCH=$(THIRDPARTY_FABRIC_CA_BRANCH) scripts/third_party_pins/fabric-ca/apply_upstream.sh

populate: populate-vendor

populate-vendor:
	@echo "Populating vendor ..."
	@dep ensure -vendor-only

populate-clean:
	rm -Rf vendor

clean:
	rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore
	rm -f integration-report.xml report.xml
	cd test/fixtures && docker-compose -f docker-compose.yaml -f docker-compose-integration-test.yaml -f docker-compose-pkcs11-test.yaml down
