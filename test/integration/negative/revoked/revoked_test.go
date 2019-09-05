/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package revoked

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"io/ioutil"

	"bytes"

	"io"

	"time"

	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	msp2 "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	org1AdminUser       = "Admin"
	org2AdminUser       = "Admin"
	org1User            = "User1"
	org2User            = "User1"
	org1                = "Org1"
	org2                = "Org2"
	ordererAdminUser    = "Admin"
	ordererOrgName      = "OrdererOrg"
	channelID           = "orgchannel"
	configFilename      = "config_test.yaml"
	pathRevokeCaRoot    = "peerOrganizations/org1.example.com/ca/"
	pathParentCert      = "peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem"
	peerCertToBeRevoked = "peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/signcerts/peer0.org1.example.com-cert.pem"
	userCertToBeRevoked = "peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
)

var CRLTestRetryOpts = retry.Opts{
	Attempts:       20,
	InitialBackoff: 1 * time.Second,
	MaxBackoff:     15 * time.Second,
	BackoffFactor:  1.5,
	RetryableCodes: retry.TestRetryableCodes,
}

// Peers used for testing
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

var msps = []string{"Org1MSP", "Org2MSP"}

//TestPeerRevoke performs peer revoke test
// step 1: generate CRL
// step 2: update MSP revocation_list in channel config
// step 3: perform revoke peer test
func TestPeerAndUserRevoke(t *testing.T) {

	var err error
	//generate CRLs for Peer & User
	crlBytes := make([][]byte, 2)
	crlBytes[0], err = generateCRL(peerCertToBeRevoked)
	require.NoError(t, err, "failed to generate CRL for", peerCertToBeRevoked)
	require.NotEmpty(t, crlBytes, "CRL is empty")

	crlBytes[1], err = generateCRL(userCertToBeRevoked)
	require.NoError(t, err, "failed to generate CRL for", userCertToBeRevoked)
	require.NotEmpty(t, crlBytes, "CRL is empty")

	//join channel and install/instantiate cc needed for later tests
	joinChannelAndInstallCC(t)

	//update revocation list in channel config
	updateRevocationList(t, crlBytes)

	//wait for config update
	waitForConfigUpdate(t)

	//test if peer has been revoked
	testRevokedPeer(t)

	//test if user1 has been revoked
	testRevokedUser(t)

	//reset revocation list in channel config for other tests
	updateRevocationList(t, nil)
}

//joinChannelAndInstallCC joins channel and install/instantiate/query 'exampleCC2'
func joinChannelAndInstallCC(t *testing.T) {

	sdk, err := fabsdk.New(config.FromFile(integration.GetConfigPath(configFilename)))
	require.NoError(t, err)
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//join channel
	joinChannel(t, sdk)

	//install & instantiate a chaincode before updating revocation list for later test
	createCC(t, sdk, "exampleCC2", "github.com/example_cc", "0", true)

	//query that chaincode to make sure everything is fine
	org2UserChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))
	queryCC(t, org2UserChannelClientContext, "exampleCC2", true, "")

}

//updateRevocationList update MSP revocation_list in channel config
func updateRevocationList(t *testing.T, crlBytes [][]byte) {

	sdk, err := fabsdk.New(config.FromFile(integration.GetConfigPath(configFilename)))
	require.NoError(t, err)
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org1AdminChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))

	ledgerClient1, err := ledger.New(org1AdminChannelClientContext)
	require.NoError(t, err)

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	require.NoError(t, err)

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	require.NoError(t, err)

	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	require.NoError(t, err)

	//create read write set for channel config update
	readSet, writeSet := prepareReadWriteSets(t, crlBytes, ledgerClient1)
	//update channel config MSP revocation lists to generated CRL
	updateChannelConfig(t, readSet, writeSet, org1ResMgmt, org1MspClient, org2MspClient)
}

//waitForConfigUpdate waits for all peer till they are updated with latest channel config
func waitForConfigUpdate(t *testing.T) {

	sdk, err := fabsdk.New(config.FromFile(integration.GetConfigPath(configFilename)))
	require.NoError(t, err)
	defer sdk.Close()

	org1AdminChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))

	ledgerClient1, err := ledger.New(org1AdminChannelClientContext)
	require.NoError(t, err)

	ctx, err := org1AdminChannelClientContext()
	require.NoError(t, err)

	ready := queryRevocationListUpdates(t, ledgerClient1, ctx.EndpointConfig(), channelID)
	require.True(t, ready, "all peers are not updated with latest channel config")
}

//testRevokedPeer performs revoke peer test
func testRevokedPeer(t *testing.T) {

	sdk1, err := fabsdk.New(config.FromFile(integration.GetConfigPath(configFilename)))
	require.NoError(t, err)
	defer sdk1.Close()

	//prepare contexts
	org1AdminClientContext := sdk1.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org1UserChannelClientContext := sdk1.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))
	org2UserChannelClientContext := sdk1.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))

	// Create chaincode package for example cc
	createCC(t, sdk1, "exampleCC", "github.com/example_cc", "1", false)

	// Load specific targets for move funds test - one of the
	//targets has its certificate revoked
	loadOrgPeers(t, org1AdminClientContext)

	//query with revoked user
	t.Log("query with revoked user - should fail with 'access denied'")
	queryCC(t, org1UserChannelClientContext, "exampleCC", false, "access denied")
	//query with valid user
	t.Log("query with valid user - should fail with 'chaincode exampleCC not found'")
	queryCC(t, org2UserChannelClientContext, "exampleCC", false, "chaincode exampleCC not found")
	//query already instantiated chaincode with revoked user
	t.Log("query already instantiated chaincode with revoked user - should fail with 'access denied'")
	queryCC(t, org1UserChannelClientContext, "exampleCC2", false, "access denied")
	//query already instantiated chaincode with valid user
	t.Log("query already instantiated chaincode with valid user - should fail with 'signature validation failed'")
	queryCC(t, org2UserChannelClientContext, "exampleCC2", false, "signature validation failed")
}

//testRevokedUser performs revoke peer test
func testRevokedUser(t *testing.T) {
	var sdk *fabsdk.FabricSDK
	var err error
	sdk, err = fabsdk.New(
		config.FromFile(integration.GetConfigPath(configFilename)),
		fabsdk.WithErrorHandler(func(ctxt fab.ClientContext, channelID string, err error) {
			if strings.Contains(err.Error(), "access denied") {
				t.Logf("Closing context after error: %s", err)
				go sdk.CloseContext(ctxt)
			}
		}),
	)
	require.NoError(t, err)
	defer sdk.Close()

	//Try User2 whose certs are not revoked, should be able to query channel config
	user2ChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))
	ledgerClient, err := ledger.New(user2ChannelContext)
	require.NoError(t, err)
	cfg, err := ledgerClient.QueryConfig(ledger.WithTargetEndpoints("peer1.org2.example.com"))
	require.NoError(t, err)
	require.NotEmpty(t, cfg)

	//Try User1 whose certs are revoked, shouldn't be able to query channel config
	user1ChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))
	_, err = ledger.New(user1ChannelContext)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "access denied"))
}

//prepareReadWriteSets prepares read write sets for channel config update
func prepareReadWriteSets(t *testing.T, crlBytes [][]byte, ledgerClient *ledger.Client) (*common.ConfigGroup, *common.ConfigGroup) {

	var readSet, writeSet *common.ConfigGroup

	chCfg, err := ledgerClient.QueryConfig(ledger.WithTargetEndpoints("peer1.org2.example.com"))
	require.NoError(t, err)

	block, err := ledgerClient.QueryBlock(chCfg.BlockNumber(), ledger.WithTargetEndpoints("peer1.org2.example.com"))
	require.NoError(t, err)

	configEnv, err := resource.CreateConfigUpdateEnvelope(block.Data.Data[0])
	require.NoError(t, err)

	configUpdate := &common.ConfigUpdate{}
	proto.Unmarshal(configEnv.ConfigUpdate, configUpdate)
	readSet = configUpdate.ReadSet

	//prepare write set
	configEnv, err = resource.CreateConfigUpdateEnvelope(block.Data.Data[0])
	require.NoError(t, err)

	configUpdate = &common.ConfigUpdate{}
	proto.Unmarshal(configEnv.ConfigUpdate, configUpdate)
	writeSet = configUpdate.ReadSet

	//change write set for MSP revocation list update
	for _, org := range msps {
		val := writeSet.Groups["Application"].Groups[org].Values["MSP"].Value

		mspCfg := &msp.MSPConfig{}
		err = proto.Unmarshal(val, mspCfg)
		require.NoError(t, err)

		fabMspCfg := &msp.FabricMSPConfig{}
		err = proto.Unmarshal(mspCfg.Config, fabMspCfg)
		require.NoError(t, err)

		if len(crlBytes) > 0 {
			//append valid crl bytes to existing revocation list
			fabMspCfg.RevocationList = append(fabMspCfg.RevocationList, crlBytes...)
		} else {
			//reset
			fabMspCfg.RevocationList = nil
		}

		fabMspBytes, err := proto.Marshal(fabMspCfg)
		require.NoError(t, err)

		mspCfg.Config = fabMspBytes

		mspBytes, err := proto.Marshal(mspCfg)
		require.NoError(t, err)

		writeSet.Groups["Application"].Groups[org].Values["MSP"].Version++
		writeSet.Groups["Application"].Groups[org].Values["MSP"].Value = mspBytes
	}

	return readSet, writeSet
}

func updateChannelConfig(t *testing.T, readSet *common.ConfigGroup, writeSet *common.ConfigGroup, resmgmtClient *resmgmt.Client, org1MspClient, org2MspClient *mspclient.Client) {

	//read block template and update read/write sets
	txBytes, err := ioutil.ReadFile(integration.GetChannelConfigTxPath("twoorgs.genesis.block"))
	require.NoError(t, err)

	block := &common.Block{}
	err = proto.Unmarshal(txBytes, block)
	require.NoError(t, err)

	configUpdateEnv, err := resource.CreateConfigUpdateEnvelope(block.Data.Data[0])
	require.NoError(t, err)

	configUpdate := &common.ConfigUpdate{}
	proto.Unmarshal(configUpdateEnv.ConfigUpdate, configUpdate)
	configUpdate.ChannelId = channelID
	configUpdate.ReadSet = readSet
	configUpdate.WriteSet = writeSet

	rawBytes, err := proto.Marshal(configUpdate)
	require.NoError(t, err)

	configUpdateEnv.ConfigUpdate = rawBytes
	configUpdateBytes, err := proto.Marshal(configUpdateEnv)
	require.NoError(t, err)

	//create config envelope
	reader := createConfigEnvelopeReader(t, block.Data.Data[0], configUpdateBytes)

	org1AdminIdentity, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	require.NoError(t, err, "failed to get org1AdminIdentity")

	org2AdminIdenity, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	require.NoError(t, err, "failed to get org2AdminIdentity")

	//perform save channel for channel config update
	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfig:     reader,
		SigningIdentities: []msp2.SigningIdentity{org1AdminIdentity, org2AdminIdenity}}
	txID, err := resmgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))

	require.Nil(t, err, "error should be nil for SaveChannel ")
	require.NotEmpty(t, txID, "transaction ID should be populated ")
}

func createConfigEnvelopeReader(t *testing.T, blockData []byte, configUpdateBytes []byte) io.Reader {
	envelope := &common.Envelope{}
	err := proto.Unmarshal(blockData, envelope)
	require.NoError(t, err)

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	require.NoError(t, err)

	payload.Data = configUpdateBytes
	payloadBytes, err := proto.Marshal(payload)
	require.NoError(t, err)

	envelope.Payload = payloadBytes
	envelopeBytes, err := proto.Marshal(envelope)
	require.NoError(t, err)

	reader := bytes.NewReader(envelopeBytes)
	return reader
}

func joinChannel(t *testing.T, sdk *fabsdk.FabricSDK) {

	joinChannelFunc := func() error {

		org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
		org2AdminClientContext := sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))

		org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
		require.NoError(t, err)

		org2ResMgmt, err := resmgmt.New(org2AdminClientContext)
		require.NoError(t, err)

		// Org1 peers join channel
		if err := org1ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
			return err
		}

		// Org2 peers join channel
		if err := org2ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
			return err
		}

		t.Log("joined channel successfully")
		return nil
	}

	//join channel
	err := joinChannelFunc()
	if err == nil {
		return
	}

	if !strings.Contains(err.Error(), "genesis block retrieval failed: Orderer Server Status Code: (404) NOT_FOUND.") {
		t.Fatalf("Failed to join channel, error : %v", err)
	}

	t.Logf("Failed to join channel due to : %v, \n Now performing save channel with orderer client and retrying", err)

	ordererClientContext := sdk.Context(fabsdk.WithUser(ordererAdminUser), fabsdk.WithOrg(ordererOrgName))

	ordererResMgmt, err := resmgmt.New(ordererClientContext)
	require.NoError(t, err)

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	require.NoError(t, err)

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	require.NoError(t, err)

	org1AdminIdentity, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	require.NoError(t, err, "failed to get org1AdminIdentity")

	org2AdminIdenity, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	require.NoError(t, err, "failed to get org2AdminIdentity")

	req := resmgmt.SaveChannelRequest{ChannelID: "orgchannel",
		ChannelConfigPath: integration.GetChannelConfigTxPath("orgchannel.tx"),
		SigningIdentities: []msp2.SigningIdentity{org1AdminIdentity, org2AdminIdenity}}
	txID, err := ordererResMgmt.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, txID, "transaction ID should be populated")

	//Try again now
	err = joinChannelFunc()
	require.NoError(t, err, "failed to join channel...")

}

func queryCC(t *testing.T, channelClientContext contextAPI.ChannelProvider, ccID string, success bool, expectedMsg string) {
	chClientOrg1User, err := channel.New(channelClientContext)
	if err != nil {
		if success {
			t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
		}
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Fatalf("Expected error: '%s' , but got '%s'", expectedMsg, err)
		}
		return
	}

	resp, err := chClientOrg1User.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCDefaultQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts), channel.WithTargetEndpoints("peer0.org1.example.com"))

	if success {
		require.NoError(t, err)
		require.NotEmpty(t, resp.Responses)
		require.NotEmpty(t, resp.Payload)
		require.Equal(t, "200", string(resp.Payload))
	} else {
		if err == nil || !strings.Contains(err.Error(), expectedMsg) {
			t.Fatalf("Expected error: '%s' , but got '%s'", expectedMsg, err)
		}
		_, ok := status.FromError(err)
		assert.True(t, ok, "Expected status error")
	}
}

func createCC(t *testing.T, sdk1 *fabsdk.FabricSDK, name, path, version string, success bool) {

	org1AdminClientContext := sdk1.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org2AdminClientContext := sdk1.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))

	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	require.NoError(t, err)

	org2ResMgmt, err := resmgmt.New(org2AdminClientContext)
	require.NoError(t, err)

	ccPkg, err := packager.NewCCPackage(path, integration.GetDeployPath())
	if err != nil {
		t.Fatal(err)
	}
	installCCReq := resmgmt.InstallCCRequest{Name: name, Path: path, Version: version, Package: ccPkg}
	// Install example cc to Org1 peers
	_, err = org1ResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}
	// Install example cc to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Set up chaincode policy to 'two-of-two msps'
	ccPolicy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
	require.NoErrorf(t, err, "Error creating cc policy with both orgs to approve")
	// Org1 resource manager will instantiate 'example_cc' on 'orgchannel'
	resp, err := org1ResMgmt.InstantiateCC(
		"orgchannel",
		resmgmt.InstantiateCCRequest{
			Name:    name,
			Path:    path,
			Version: version,
			Args:    integration.ExampleCCInitArgs(),
			Policy:  ccPolicy,
		},
		resmgmt.WithTargetEndpoints("peer0.org1.example.com", "peer0.org2.example.com"),
	)

	if success {
		require.NoError(t, err)
		require.NotEmpty(t, resp.TransactionID)
	} else {
		require.Errorf(t, err, "Expecting error instantiating CC on peer with revoked certificate")
		stat, ok := status.FromError(err)
		require.Truef(t, ok, "Expecting error to be a status error, but got ", err)
		require.Equalf(t, stat.Code, int32(status.SignatureVerificationFailed), "Expecting signature verification error due to revoked cert, but got", err)
		require.Truef(t, strings.Contains(err.Error(), "the creator certificate is not valid"), "Expecting error message to contain 'the creator certificate is not valid' but got", err)
	}
}

func loadOrgPeers(t *testing.T, ctxProvider contextAPI.ClientProvider) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("context creation failed: %s", err)
	}

	org1Peers, ok := ctx.EndpointConfig().PeersConfig(org1)
	assert.True(t, ok)

	org2Peers, ok := ctx.EndpointConfig().PeersConfig(org2)
	assert.True(t, ok)

	orgTestPeer0, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org1Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org2Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

}

func queryRevocationListUpdates(t *testing.T, client *ledger.Client, config fab.EndpointConfig, chID string) bool {
	installed, err := retry.NewInvoker(retry.New(CRLTestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok := isChannelConfigUpdated(t, client, config, chID)
			if !ok {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), "Revocation list is not updated in all peers", nil)
			}
			return &ok, nil
		},
	)

	require.NoErrorf(t, err, "Got error checking if chaincode was installed")
	return *(installed).(*bool)
}

func isChannelConfigUpdated(t *testing.T, client *ledger.Client, config fab.EndpointConfig, chID string) bool {
	chPeers := config.ChannelPeers(chID)
	t.Logf("Performing config update check on %d channel peers in channel '%s'", len(chPeers), chID)
	updated := len(chPeers) > 0
	for _, chPeer := range chPeers {
		t.Logf("waiting for [%s] msp update", chPeer.URL)
		chCfg, err := client.QueryConfig(ledger.WithTargetEndpoints(chPeer.URL))
		if err != nil || len(chCfg.MSPs()) == 0 {
			return false
		}
		for _, mspCfg := range chCfg.MSPs() {
			fabMspCfg := &msp.FabricMSPConfig{}
			err = proto.Unmarshal(mspCfg.Config, fabMspCfg)
			if err != nil {
				return false
			}
			if fabMspCfg.Name == "OrdererMSP" {
				continue
			}
			t.Logf("length of revocation list found in peer[%s] is %d", chPeer.URL, len(fabMspCfg.RevocationList))
			updated = updated && len(fabMspCfg.RevocationList) > 1
		}
	}
	t.Logf("check result :%v \n\n", updated)
	return updated
}

func generateCRL(cerPath string) ([]byte, error) {

	root := integration.GetCryptoConfigPath(pathRevokeCaRoot)
	var parentKey string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "_sk") {
			parentKey = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	key, err := loadPrivateKey(parentKey)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load private key")
	}

	cert, err := loadCert(integration.GetCryptoConfigPath(pathParentCert))
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load cert")
	}

	certToBeRevoked, err := loadCert(integration.GetCryptoConfigPath(cerPath))
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load cert")
	}

	crlBytes, err := revokeCert(certToBeRevoked, cert, key)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to revoke cert")
	}

	return crlBytes, nil
}

func loadPrivateKey(path string) (interface{}, error) {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key, err := utils.PEMtoPrivateKey(raw, []byte(""))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func loadCert(path string) (*x509.Certificate, error) {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}

func revokeCert(certToBeRevoked *x509.Certificate, parentCert *x509.Certificate, parentKey interface{}) ([]byte, error) {

	//Create a revocation record for the user
	clientRevocation := pkix.RevokedCertificate{
		SerialNumber:   certToBeRevoked.SerialNumber,
		RevocationTime: time.Now().UTC(),
	}

	curRevokedCertificates := []pkix.RevokedCertificate{clientRevocation}
	//Generate new CRL that includes the user's revocation
	newCrlList, err := parentCert.CreateCRL(rand.Reader, parentKey, curRevokedCertificates, time.Now().UTC(), time.Now().UTC().AddDate(20, 0, 0))
	if err != nil {
		return nil, err
	}

	//CRL pem Block
	crlPemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: newCrlList,
	}
	var crlBuffer bytes.Buffer
	//Encode it to X509 CRL pem format print it out
	err = pem.Encode(&crlBuffer, crlPemBlock)
	if err != nil {
		return nil, err
	}

	return crlBuffer.Bytes(), nil
}
