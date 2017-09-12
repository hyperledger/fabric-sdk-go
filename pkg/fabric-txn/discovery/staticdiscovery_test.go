/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestDiscovery(t *testing.T) {

	invalidChannel, err := setupChannel("not-configured")
	if err != nil {
		t.Fatalf("Failed to setup channel: %s", err)
	}

	validChannel, err := setupChannel("mychannel")
	if err != nil {
		t.Fatalf("Failed to setup channel: %s", err)
	}

	config, err := config.InitConfig("../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}

	discoveryProvider, err := NewDiscoveryProvider(config)
	if err != nil {
		t.Fatalf("Failed to  setup discovery provider: %s", err)
	}

	discoveryService, err := discoveryProvider.NewDiscoveryService(invalidChannel)
	if err == nil {
		t.Fatalf("Should have failed to setup discovery service for non-configured channel")
	}

	discoveryService, err = discoveryProvider.NewDiscoveryService(validChannel)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	peers, err := discoveryService.GetPeers("testCC")
	if err != nil {
		t.Fatalf("Failed to get peers from discovery service: %s", err)
	}

	expectedNumOfPeeers := 1
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}

}

func setupChannel(name string) (*channel.Channel, error) {
	client := setupTestClient()
	return channel.NewChannel(name, client)
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return client
}
