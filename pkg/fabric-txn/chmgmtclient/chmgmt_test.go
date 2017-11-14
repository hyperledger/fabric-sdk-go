/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chmgmtclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

const channelConfig = "./testdata/test.tx"
const networkCfg = "../../../test/fixtures/config/config_test.yaml"

func TestSaveChannel(t *testing.T) {

	cc := setupChannelMgmtClient(t)

	// Test empty channel request
	err := cc.SaveChannel(chmgmtclient.SaveChannelRequest{})
	if err == nil {
		t.Fatalf("Should have failed for empty channel request")
	}

	// Test empty channel name
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatalf("Should have failed for empty channel id")
	}

	// Test empty channel config
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: ""})
	if err == nil {
		t.Fatalf("Should have failed for empty channel config")
	}

	// Test extract configuration error
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: "./testdata/extractcherr.tx"})
	if err == nil {
		t.Fatalf("Should have failed to extract configuration")
	}

	// Test sign channel error
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: "./testdata/signcherr.tx"})
	if err == nil {
		t.Fatalf("Should have failed to sign configuration")
	}

	// Test valid Save Channel request (success)
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err != nil {
		t.Fatal(err)
	}

}

func TestSaveChannelFailure(t *testing.T) {

	// Set up client with error in create channel
	errClient := fcmocks.NewMockInvalidClient()
	user := fcmocks.NewMockUser("test")
	errClient.SetUserContext(user)
	network := getNetworkConfig(t)

	cc, err := NewChannelMgmtClient(errClient, network)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	// Test create channel failure
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatal("Should have failed with create channel error")
	}

}

func TestNoSigningUserFailure(t *testing.T) {

	// Setup client without user context
	client := fcmocks.NewMockClient()
	network := getNetworkConfig(t)

	cc, err := NewChannelMgmtClient(client, network)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	// Test save channel without signing user set (and no default context user)
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatal("Should have failed due to missing signing user")
	}

}

func TestSaveChannelWithOpts(t *testing.T) {

	cc := setupChannelMgmtClient(t)

	// Valid request (same for all options)
	req := chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig}

	// Test empty option (default order is random orderer from config)
	opts := chmgmtclient.SaveChannelOpts{}
	err := cc.SaveChannelWithOpts(req, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test valid orderer ID
	opts.OrdererID = "orderer.example.com"
	err = cc.SaveChannelWithOpts(req, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test invalid orderer ID
	opts.OrdererID = "Invalid"
	err = cc.SaveChannelWithOpts(req, opts)
	if err == nil {
		t.Fatal("Should have failed for invalid orderer ID")
	}
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

func getNetworkConfig(t *testing.T) *config.Config {
	config, err := config.InitConfig(networkCfg)
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setupChannelMgmtClient(t *testing.T) *ChannelMgmtClient {

	fcClient := setupTestClient()
	network := getNetworkConfig(t)

	consClient, err := NewChannelMgmtClient(fcClient, network)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	return consClient
}
