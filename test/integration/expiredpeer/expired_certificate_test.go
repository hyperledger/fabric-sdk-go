/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package expiredpeer

import (
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
)

const (
	org1             = "Org1"
	org2             = "Org2"
	ordererAdminUser = "Admin"
	ordererOrgName   = "ordererorg"
	org1AdminUser    = "Admin"
	org2AdminUser    = "Admin"
	configPath       = "../../fixtures/config/config_test.yaml"
	expiredCertPath  = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/expiredtlsca/expired.pem"
)

// TestExpiredPeersCert - peer0.org1.example.com was configured with expired certificate
func TestExpiredPeersCert(t *testing.T) {

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(getConfigBackend(t))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//prepare contexts
	ordererClientContext := sdk.Context(fabsdk.WithUser(ordererAdminUser), fabsdk.WithOrg(ordererOrgName))
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(ordererClientContext)
	if err != nil {
		t.Fatal(err)
	}

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	if err != nil {
		t.Fatalf("failed to create org1MspClient, err : %s", err)
	}

	// Get signing identity that is used to sign create channel request
	org1AdminUser, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	if err != nil {
		t.Fatalf("failed to get org1AdminUser, err : %s", err)
	}

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	if err != nil {
		t.Fatalf("failed to create org2MspClient, err : %s", err)
	}

	org2AdminUser, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	if err != nil {
		t.Fatalf("failed to get org2AdminUser, err : %s", err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: "orgchannel",
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannel.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser, org2AdminUser}}
	txID, err := chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, txID, "transaction ID should be populated")

	// Org1 resource management client (Org1 is default org)
	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}
	// Org1 peers join channel
	err = org1ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err == nil {
		//full error from GRPC log 'Failed to dial peer0.org1.example.com:7051: connection error: desc = "transport: authentication handshake failed: x509: certificate has expiredorderer or is not yet valid"; please retry.'
		t.Fatal("Expected error: 'Error join channel failed: SendProposal failed...")
	}

}

func getConfigBackend(t *testing.T) core.ConfigProvider {

	return func() ([]core.ConfigBackend, error) {
		configBackends, err := config.FromFile(configPath)()
		assert.Nil(t, err, "failed to read config backend from file", err)
		backendMap := make(map[string]interface{})

		networkConfig := endpointConfigEntity{}
		//get valid peer config
		err = lookup.New(configBackends...).UnmarshalKey("peers", &networkConfig.Peers)
		assert.Nil(t, err, "failed to unmarshal peer network config")
		//change cert path to expired one
		peer1 := networkConfig.Peers["peer0.org1.example.com"]
		peer1.TLSCACerts.Path = expiredCertPath
		networkConfig.Peers["peer0.org1.example.com"] = peer1
		backendMap["peers"] = networkConfig.Peers

		backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
		return append(backends, configBackends...), nil
	}
}

//endpointConfigEntity contains endpoint config elements needed by endpointconfig
type endpointConfigEntity struct {
	Peers map[string]fabImpl.PeerConfig
}
