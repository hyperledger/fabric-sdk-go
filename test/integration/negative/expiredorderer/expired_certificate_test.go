/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package expiredorderer

import (
	"os"
	"testing"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"google.golang.org/grpc/grpclog"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/assert"
)

const (
	org1             = "Org1"
	org2             = "Org2"
	ordererAdminUser = "Admin"
	ordererOrgName   = "OrdererOrg"
	org1AdminUser    = "Admin"
	org2AdminUser    = "Admin"
	configFilename   = "config_test.yaml"
	expiredCertPath  = "${FABRIC_SDK_GO_PROJECT_PATH}/test/integration/negative/testdata/ordererOrganizations/example.com/expiredtlsca/expired.pem"
)

var logger = logging.NewLogger("test-logger")

// TestExpiredCert
func TestExpiredCert(t *testing.T) {
	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")
	grpclog.SetLogger(logger)

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
		ChannelConfigPath: integration.GetChannelConfigTxPath("orgchannel.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser, org2AdminUser}}
	_, err = chMgmtClient.SaveChannel(req)
	//error in GRPC log is ' Failed to dial orderer.example.com:7050: connection error: desc = "transport: authentication handshake failed: x509: certificate has expiredorderer or is not yet valid"; '
	if err == nil {
		t.Fatal("Expected error: calling orderer 'orderer.example.com:7050' failed: Orderer Client Status Code: (2) CONNECTION_FAILED....")
	}

}

func getConfigBackend(t *testing.T) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
		configBackends, err := config.FromFile(integration.GetConfigPath(configFilename))()
		assert.Nil(t, err, "failed to read config backend from file, %s", err)
		backendMap := make(map[string]interface{})

		networkConfig := endpointConfigEntity{}
		//get valid orderers config
		err = lookup.New(configBackends...).UnmarshalKey("orderers", &networkConfig.Orderers)
		assert.Nil(t, err, "failed to unmarshal peer network config")

		//change cert path to expired one
		orderer1 := networkConfig.Orderers["orderer.example.com"]
		orderer1.TLSCACerts.Path = expiredCertPath
		networkConfig.Orderers["orderer.example.com"] = orderer1
		backendMap["orderers"] = networkConfig.Orderers

		backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
		return append(backends, configBackends...), nil
	}
}

//endpointConfigEntity contains endpoint config elements needed by endpointconfig
type endpointConfigEntity struct {
	Orderers map[string]fabImpl.OrdererConfig
}
