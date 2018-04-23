/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package expiredorderer

import (
	"os"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"google.golang.org/grpc/grpclog"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
)

const (
	org1             = "Org1"
	org2             = "Org2"
	ordererAdminUser = "Admin"
	ordererOrgName   = "ordererorg"
	org1AdminUser    = "Admin"
	org2AdminUser    = "Admin"
	configPath       = "../../fixtures/config/config_test.yaml"
	expiredCertPath  = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/ordererOrganizations/example.com/expiredtlsca/expired.pem"
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
	time.Sleep(100 * time.Millisecond)

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

	// Get signing identity that is used to sign create channel request
	org1AdminUser, err := integration.GetSigningIdentity(sdk, org1AdminUser, org1)
	if err != nil {
		t.Fatalf("failed to get org1AdminUser, err : %v", err)
	}

	org2AdminUser, err := integration.GetSigningIdentity(sdk, org2AdminUser, org2)
	if err != nil {
		t.Fatalf("failed to get org2AdminUser, err : %v", err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: "orgchannel",
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannel.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser, org2AdminUser}}
	_, err = chMgmtClient.SaveChannel(req)
	//error in GRPC log is ' Failed to dial orderer.example.com:7050: connection error: desc = "transport: authentication handshake failed: x509: certificate has expiredorderer or is not yet valid"; '
	if err == nil {
		t.Fatalf("Expected error: calling orderer 'orderer.example.com:7050' failed: Orderer Client Status Code: (2) CONNECTION_FAILED....")
	}
	time.Sleep(100 * time.Millisecond)

}

func getConfigBackend(t *testing.T) core.ConfigProvider {

	return func() (core.ConfigBackend, error) {
		backend, err := config.FromFile(configPath)()
		if err != nil {
			t.Fatalf("failed to read config backend from file, %v", err)
		}
		backendMap := make(map[string]interface{})

		networkConfig := fab.NetworkConfig{}
		//get valid orderers config
		err = lookup.New(backend).UnmarshalKey("orderers", &networkConfig.Orderers)
		if err != nil {
			t.Fatalf("failed to unmarshal peer network config, %v", err)
		}
		//change cert path to expired one
		orderer1 := networkConfig.Orderers["orderer.example.com"]
		orderer1.TLSCACerts.Path = expiredCertPath
		networkConfig.Orderers["orderer.example.com"] = orderer1
		backendMap["orderers"] = networkConfig.Orderers

		return &mocks.MockConfigBackend{KeyValueMap: backendMap, CustomBackend: backend}, nil
	}
}
