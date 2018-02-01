/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestChannelConfigs(t *testing.T) {

	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)

	channel, _ := New(ctx, mocks.NewMockChannelCfg("testChannel"))

	if channel.IsReadonly() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("Is Readonly test failed")
	}

	if channel.UpdateChannel() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("UpdateChannel test failed")
	}

	channel.SetMSPManager(nil)

}
