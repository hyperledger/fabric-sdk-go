#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# This MakeFile assumes that fabric, and fabric-ca were cloned and their docker
# images were created using the make docker command in the respective directories
#
# Supported Targets:
# all : runs unit and integration tests
# depend: installs test dependencies
# unit-test: runs all the unit tests
# integration-test: runs all the integration tests
# checks: runs all check conditions (license, spelling, linting)
# clean: stops docker conatainers used for integration testing
#

all: checks unit-test integration-test

depend:
	@test/scripts/dependencies.sh

checks: depend license lint spelling

.PHONY: license
license:
	@test/scripts/check_license.sh

lint:
	@test/scripts/check_lint.sh

spelling:
	@test/scripts/check_spelling.sh

unit-test: checks
	@test/scripts/unit.sh

unit-tests: unit-test

integration-test: clean depend
	@test/scripts/integration.sh

integration-tests: integration-test

clean:
	rm -Rf /tmp/enroll_user /tmp/msp /tmp/keyvaluestore
	rm -f integration-report.xml report.xml
	cd test/fixtures && docker-compose down
