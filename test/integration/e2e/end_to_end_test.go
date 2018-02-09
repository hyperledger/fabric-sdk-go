/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"testing"
)

func TestE2E(t *testing.T) {
	//End to End testing
	runWithConfigFixture(t)

	//Using setup done set above by end to end test, run below test with new config which has no orderer config inside
	runWithNoOrdererConfigFixture(t)
}
