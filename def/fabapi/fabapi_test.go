/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import "testing"

func TestNewDefaultSDK(t *testing.T) {

	setup := Options{
		ConfigFile: "../../test/fixtures/config/config_test.yaml",
		OrgID:      "org1",
		StateStoreOpts: StateStoreOpts{
			Path: "/tmp/state",
		},
	}

	_, err := NewSDK(setup)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}
