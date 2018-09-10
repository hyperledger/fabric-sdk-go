// +build prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"testing"
)

//TestMultiOrgWithSingleOrgConfig cannot be run in prev tests since fabric version greater than v1.1
// supports dynamic discovery
func TestMultiOrgWithSingleOrgConfig(t *testing.T, examplecc string) {
	//test nothing
	t.Logf("Dynamic discovery tests didn't run for '%s', since tests are running prev-release", examplecc)
}

// DistributedSignaturesTests is not supported in prev tests
func DistributedSignaturesTests(t *testing.T, examplecc string) {
	t.Logf("Distributed Signatures tests require Dynamic Discovery which is not available in prev-release")
}
