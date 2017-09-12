/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

func TestNewDefaultSDK(t *testing.T) {

	setup := Options{
		ConfigFile: "../../test/fixtures/config/config_test.yaml",
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/state",
		},
	}

	sdk, err := NewSDK(setup)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	_, err = sdk.NewChannelClient("mychannel", "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

}
