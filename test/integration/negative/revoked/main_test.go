/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package revoked

import (
	"fmt"
	"os"
	"testing"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
)

const (
	org1           = "Org1"
	org2           = "Org2"
	channelID      = "orgchannel"
	configFilename = "config_test.yaml"
)

// SDK
var sdk *fabsdk.FabricSDK

// Org MSP clients
var org1MspClient *mspclient.Client
var org2MspClient *mspclient.Client

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		panic(fmt.Sprintf("unable to setup [%s]", err))
	}
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setup() error {
	// Create SDK setup for the integration tests
	var err error
	sdk, err = fabsdk.New(getConfigBackend())
	if err != nil {
		return errors.Wrap(err, "Failed to create new SDK")
	}

	org1MspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	if err != nil {
		return errors.Wrap(err, "failed to create org1MspClient")
	}

	org2MspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	if err != nil {
		return errors.Wrap(err, "failed to create org2MspClient")
	}

	return nil
}

func teardown() {
	if sdk != nil {
		sdk.Close()
	}
}

//configOverride to override existing config backend
type configOverride struct {
	Client        fabImpl.ClientConfig
	Channels      map[string]fabImpl.ChannelEndpointConfig
	Organizations map[string]fabImpl.OrganizationConfig
	Orderers      map[string]fabImpl.OrdererConfig
	Peers         map[string]fabImpl.PeerConfig
}

func getConfigBackend() core.ConfigProvider {

	return func() ([]core.ConfigBackend, error) {
		configBackends, err := config.FromFile(integration.GetConfigPath(configFilename))()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read config backend from file, %v")
		}
		backendMap := make(map[string]interface{})

		networkConfig := configOverride{}
		//get valid peer config
		err = lookup.New(configBackends...).UnmarshalKey("peers", &networkConfig.Peers)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal peer network config, %v")
		}

		//customize peer0.org2 to peer1.org2
		peer2 := networkConfig.Peers["peer0.org2.example.com"]
		peer2.URL = "peer1.org2.example.com:9051"
		peer2.GRPCOptions["ssl-target-name-override"] = "peer1.org2.example.com"

		//remove peer0.org2
		delete(networkConfig.Peers, "peer0.org2.example.com")

		//add peer1.org2
		networkConfig.Peers["peer1.org2.example.com"] = peer2

		//get valid org2
		err = lookup.New(configBackends...).UnmarshalKey("organizations", &networkConfig.Organizations)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal organizations network config, %v")
		}

		//Customize org2
		org2 := networkConfig.Organizations["org2"]
		org2.Peers = []string{"peer1.org2.example.com"}
		org2.MSPID = "Org2MSP"
		networkConfig.Organizations["org2"] = org2

		//custom channel
		err = lookup.New(configBackends...).UnmarshalKey("channels", &networkConfig.Channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entityMatchers network config, %v")
		}

		orgChannel := networkConfig.Channels[channelID]
		delete(orgChannel.Peers, "peer0.org2.example.com")
		orgChannel.Peers["peer1.org2.example.com"] = fabImpl.PeerChannelConfig{
			EndorsingPeer:  true,
			ChaincodeQuery: true,
			LedgerQuery:    true,
			EventSource:    false,
		}
		networkConfig.Channels[channelID] = orgChannel

		//Customize backend with update peers, organizations, channels and entity matchers config
		backendMap["peers"] = networkConfig.Peers
		backendMap["organizations"] = networkConfig.Organizations
		backendMap["channels"] = networkConfig.Channels

		backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
		return append(backends, configBackends...), nil
	}
}
