// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package orgs

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dsChannelSDK       = "dschannelsdk"
	dsChannelExternal  = "dschannelext"
	adminUser          = "Admin"
	user1              = "User1"
	mainConfigFilename = "config_e2e_multiorg_bootstrap.yaml"
)

type dsClientCtx struct {
	org   string
	sdk   *fabsdk.FabricSDK
	clCtx contextApi.ClientProvider
	rsCl  *resmgmt.Client
}

type chCfgSignatures struct {
	org1DsChannelCfgSig    *common.ConfigSignature
	org2DsChannelCfgSig    *common.ConfigSignature
	org1MSPDsChannelCfgSig *common.ConfigSignature
	org2MSPDsChannelCfgSig *common.ConfigSignature
}

// DistributedSignaturesTests will create at least 2 clients, each from 2 different orgs and creates two channel where these 2 orgs are members
// one channel created by using the conventional SDK signatures (exported into a file and loaded to simulate external signature loading)
// the second one is created by using OpenSSL to sign the channel Config data.
func DistributedSignaturesTests(t *testing.T, examplecc string) {
	ordererClCtx := createDSClientCtx(t, ordererOrgName)
	defer ordererClCtx.sdk.Close()

	org1ClCtx := createDSClientCtx(t, org1)
	defer org1ClCtx.sdk.Close()

	org2ClCtx := createDSClientCtx(t, org2)
	defer org2ClCtx.sdk.Close()

	// use SDK signing
	e2eCreateAndQueryChannel(t, ordererClCtx, org1ClCtx, org2ClCtx, dsChannelSDK, exampleCC)

	if isOpensslAvailable(t) {
		// use OpenSSL signing
		e2eCreateAndQueryChannel(t, ordererClCtx, org1ClCtx, org2ClCtx, dsChannelExternal, exampleCC)
	}

}

func e2eCreateAndQueryChannel(t *testing.T, ordererClCtx, org1ClCtx, org2ClCtx *dsClientCtx, channelID, examplecc string) {
	sigDir, err := ioutil.TempDir("", channelID)
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() {
		err = os.RemoveAll(sigDir)
		require.NoError(t, err, "Failed to remove temporary directory")
	}()

	t.Logf("created tempDir for %s signatures: %s", channelID, sigDir)
	chConfigPath := integration.GetChannelConfigPath(fmt.Sprintf("%s.tx", channelID))
	chConfigOrg1MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", channelID, org1))
	chConfigOrg2MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", channelID, org2))
	isSDKSigning := channelID == dsChannelSDK
	sigs := generateSignatures(t, org1ClCtx, org2ClCtx, chConfigPath, chConfigOrg1MSPPath, chConfigOrg2MSPPath, sigDir, isSDKSigning)
	saveChannel(t, ordererClCtx, org1ClCtx, org2ClCtx, channelID, chConfigPath, chConfigOrg1MSPPath, chConfigOrg2MSPPath, sigs, true)
	// Org1 peers join channel
	err = org1ClCtx.rsCl.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "Org1 peers failed to JoinChannel")

	// Org2 peers join channel
	err = org2ClCtx.rsCl.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "Org2 peers failed to JoinChannel")

	// Ensure that Gossip has propagated it's view of local peers before invoking
	// install since some peers may be missed if we call InstallCC too early
	org1Peers, err := integration.DiscoverLocalPeers(org1ClCtx.clCtx, 2)
	require.NoError(t, err)
	org2Peers, err := integration.DiscoverLocalPeers(org2ClCtx.clCtx, 2)
	require.NoError(t, err)

	ccVersion := "1" // ccVersion= 1 because previous test increased the ccVersion # on the peers.

	// instantiate example_CC on dschannel
	instantiateCC(t, org1ClCtx.rsCl, exampleCC, ccVersion, channelID)

	// Ensure the CC is instantiated on all peers in both orgs
	found := queryInstantiatedCC(t, org1, org1ClCtx.rsCl, channelID, exampleCC, ccVersion, org1Peers)
	require.True(t, found, "Failed to find instantiated chaincode [%s:%s] in at least one peer in Org1 on channel [%s]", exampleCC, ccVersion, channelID)

	found = queryInstantiatedCC(t, org2, org2ClCtx.rsCl, channelID, exampleCC, ccVersion, org2Peers)
	require.True(t, found, "Failed to find instantiated chaincode [%s:%s] in at least one peer in Org2 on channel [%s]", exampleCC, ccVersion, channelID)

	// test regular querying on dschannel from org1 and org2
	testQueryingOrgs(t, org1ClCtx.sdk, org2ClCtx.sdk, channelID, examplecc)
}

func generateSignatures(t *testing.T, org1ClCtx, org2ClCtx *dsClientCtx, chConfigPath, chConfigOrg1MSPPath, chConfigOrg2MSPPath, sigDir string, isSDKSigning bool) chCfgSignatures {
	chCfgSigs := &chCfgSignatures{}

	// create org1 ConfigSignature
	chCfgSigs.org1DsChannelCfgSig = executeSigning(t, org1ClCtx, chConfigPath, adminUser, sigDir, isSDKSigning)
	//t.Logf("org1DsChannelCfgSig:[%+v]", chCfgSigs.org1DsChannelCfgSig)

	// create org2 ConfigSignature
	chCfgSigs.org2DsChannelCfgSig = executeSigning(t, org2ClCtx, chConfigPath, adminUser, sigDir, isSDKSigning)
	//t.Logf("org2DsChannelCfgSig:[%+v]", chCfgSigs.org2DsChannelCfgSig)

	// create signature for anchor peer of org1
	chCfgSigs.org1MSPDsChannelCfgSig = executeSigning(t, org1ClCtx, chConfigOrg1MSPPath, adminUser, sigDir, isSDKSigning)
	//t.Logf("org1MSPDsChannelCfgSig:[%+v]", chCfgSigs.org1MSPDsChannelCfgSig)

	// create signature for anchor peer of org2
	chCfgSigs.org2MSPDsChannelCfgSig = executeSigning(t, org2ClCtx, chConfigOrg2MSPPath, adminUser, sigDir, isSDKSigning)
	//t.Logf("org2MSPDsChannelCfgSig:[%+v]", chCfgSigs.org2MSPDsChannelCfgSig)

	return *chCfgSigs
}

func saveChannel(t *testing.T, ordererClCtx, org1ClCtx, org2ClCtx *dsClientCtx, channelID, chConfigPath, chConfigOrg1MSPPath, chConfigOrg2MSPPath string, sigs chCfgSignatures, isGenesis bool) {
	// create channel on the orderer
	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfigPath: chConfigPath}
	txID, err := ordererClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(sigs.org1DsChannelCfgSig, sigs.org2DsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", channelID, ordererOrgName)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", ordererOrgName)

	var lastConfigBlock uint64
	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, channelID, isGenesis, lastConfigBlock)

	// create channel on anchor peer of org1
	req.ChannelConfigPath = chConfigOrg1MSPPath
	txID, err = org1ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(sigs.org1MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s for anchor peer of %s", channelID, org1)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for anchor peer of %s", org1)

	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, channelID, false, lastConfigBlock)

	// create channel on anchor peer of org2
	req.ChannelConfigPath = chConfigOrg2MSPPath
	txID, err = org2ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(sigs.org2MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s for anchor peer of %s", channelID, org2)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for anchor peer of %s", org2)

	integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, channelID, false, lastConfigBlock)
}

func executeSigning(t *testing.T, dsCtx *dsClientCtx, chConfigPath, user, sigDir string, isSDKSigning bool) *common.ConfigSignature {
	if isSDKSigning {
		return executeSDKSigning(t, dsCtx, chConfigPath, user, sigDir)
	}
	return executeExternalSigning(t, dsCtx, chConfigPath, user, sigDir)
}

func executeSDKSigning(t *testing.T, dsCtx *dsClientCtx, chConfigPath, user, sigDir string) *common.ConfigSignature {
	chCfgName := getBaseChCfgFileName(chConfigPath)

	channelCfgSigSDK := createSignatureFromSDK(t, dsCtx, chConfigPath, user)
	f, err := os.Create(path.Join(sigDir, fmt.Sprintf("%s_%s_%s_sbytes.txt.sha256", chCfgName, dsCtx.org, user)))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = f.Close()
		require.NoError(t, err, "Failed to close signature file")
	}()
	bufferedWriter := bufio.NewWriter(f)
	_, err = bufferedWriter.Write(channelCfgSigSDK.Signature)
	assert.NoError(t, err, "must be able to write signature of [%s-%s] to buffer", dsCtx.org, user)
	err = bufferedWriter.Flush()
	assert.NoError(t, err, "must be able to flush signature header of [%s-%s] from buffer", dsCtx.org, user)
	shf, err := os.Create(path.Join(sigDir, fmt.Sprintf("%s_%s_%s_sHeaderbytes.txt", chCfgName, dsCtx.org, user)))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = shf.Close()
		require.NoError(t, err, "Failed to close signature header file")
	}()
	bufferedWriter = bufio.NewWriter(shf)
	_, err = bufferedWriter.Write(channelCfgSigSDK.SignatureHeader)
	assert.NoError(t, err, "must be able to write signature header of [%s-%s] to buffer", dsCtx.org, user)
	err = bufferedWriter.Flush()
	assert.NoError(t, err, "must be able to flush signature header of [%s-%s] from buffer", dsCtx.org, user)
	// now that signature is stored in the filesystem, load it to simulate external signature read
	return loadExternalSignature(t, dsCtx.org, chConfigPath, user, sigDir)
}

func getBaseChCfgFileName(chConfigPath string) string {
	chCfgName := filepath.Base(chConfigPath)
	chCfgName = strings.TrimSuffix(chCfgName, filepath.Ext(chCfgName))
	return chCfgName
}

func createSignatureFromSDK(t *testing.T, dsCtx *dsClientCtx, chConfigPath string, user string) *common.ConfigSignature {
	mspClient, err := mspclient.New(dsCtx.sdk.Context(), mspclient.WithOrg(dsCtx.org))
	require.NoError(t, err, "error creating a new msp management client for %s", dsCtx.org)
	usr, err := mspClient.GetSigningIdentity(user)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", dsCtx.org)

	signature, err := dsCtx.rsCl.CreateConfigSignature(usr, chConfigPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org1)

	return signature
}

func executeExternalSigning(t *testing.T, clCtx *dsClientCtx, chConfigPath, user string, sigDir string) *common.ConfigSignature {
	// example generating and loading an external signature (not signed by the SDK)
	generateChConfigData(t, clCtx, chConfigPath, user, sigDir)

	// sign signature data with external tool (script running openssl)
	generateExternalChConfigSignature(t, clCtx.org, user, chConfigPath, sigDir)

	cs := loadExternalSignature(t, clCtx.org, chConfigPath, user, sigDir)

	return cs
}

func createDSClientCtx(t *testing.T, org string) *dsClientCtx {
	if org == ordererOrgName {
		return createOrderDsClientCtx(t)
	}

	d := &dsClientCtx{org: org}

	var err error
	b := getCustomConfigBackend(t, org)
	if integration.IsLocal() {
		//If it is a local test then add entity mapping to config backend to parse URLs
		b = integration.AddLocalEntityMapping(b)
	}

	// create SDK with dynamic discovery enabled
	d.sdk, err = fabsdk.New(b, fabsdk.WithServicePkg(&DynDiscoveryProviderFactory{}))
	require.NoError(t, err, "error creating a new sdk client for %s", org)

	d.clCtx = d.sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org))
	d.rsCl, err = resmgmt.New(d.clCtx)
	require.NoError(t, err, "error creating a new resource management client for %s", org)
	return d
}

func createOrderDsClientCtx(t *testing.T) *dsClientCtx {
	sdk, err := fabsdk.New(integration.ConfigBackend)
	require.NoError(t, err, "error creating a new sdk client for %s", ordererOrgName)

	ordererCtx := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(ordererOrgName))

	// create Channel management client for OrdererOrg
	chMgmtClient, err := resmgmt.New(ordererCtx)
	require.NoError(t, err, "error creating a new resource management client for %s", ordererOrgName)

	return &dsClientCtx{
		org:   ordererOrgName,
		sdk:   sdk,
		clCtx: ordererCtx,
		rsCl:  chMgmtClient,
	}
}

func getCustomConfigBackend(t *testing.T, org string) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
		configBackends, err := config.FromFile(integration.GetConfigPath(mainConfigFilename))()
		require.NoError(t, err, "failed to read config backend from file for org %s", org)

		configBackendsOverrides := getOrgBackendsOverride(configBackends...)

		// change org name and tls path with 'org' name and tls path
		clBackend := clientCfg
		clBackend = strings.Replace(clBackend, "organization: org1", fmt.Sprintf("organization: %s", strings.ToLower(org)), -1)
		clBackend = strings.Replace(clBackend, "tls.example.com", fmt.Sprintf("%s.example.com", strings.ToLower(org)), -1)
		r := bytes.NewReader([]byte(clBackend))
		clientBackend := config.FromReader(r, "yaml")
		cb, err := clientBackend()
		require.NoError(t, err, "Failed to create new customer backend config")

		backends := append(cb, configBackendsOverrides)

		return append(backends, configBackends...), nil

	}
}

func getOrgBackendsOverride(backend ...core.ConfigBackend) *mocks.MockConfigBackend {
	//Create dschannelsdk and dschannelext channels
	networkConfig := endpointConfigEntity{}

	err := lookup.New(backend...).UnmarshalKey("channels", &networkConfig.Channels)
	if err != nil {
		panic(err)
	}

	// fetch existing channel in config
	mychannel := networkConfig.Channels["orgchannel"]

	// add both dschannelsdk and dschannelext
	networkConfig.Channels[dsChannelSDK] = mychannel
	networkConfig.Channels[dsChannelExternal] = mychannel

	mockConfigBackend := getCustomBackend(backend...)
	mockConfigBackend.KeyValueMap["channels"] = networkConfig.Channels

	return mockConfigBackend
}

func getCustomBackend(backend ...core.ConfigBackend) *mocks.MockConfigBackend {

	backendMap := make(map[string]interface{})
	backendMap["client"], _ = backend[0].Lookup("client")
	backendMap["certificateAuthorities"], _ = backend[0].Lookup("certificateAuthorities")
	backendMap["entityMatchers"], _ = backend[0].Lookup("entityMatchers")
	backendMap["peers"], _ = backend[0].Lookup("peers")
	backendMap["organizations"], _ = backend[0].Lookup("organizations")
	backendMap["orderers"], _ = backend[0].Lookup("orderers")
	backendMap["channels"], _ = backend[0].Lookup("channels")

	return &mocks.MockConfigBackend{KeyValueMap: backendMap}
}

func testQueryingOrgs(t *testing.T, org1sdk *fabsdk.FabricSDK, org2sdk *fabsdk.FabricSDK, dsChannel, examplecc string) {
	//prepare context
	org1ChannelClientContext := org1sdk.ChannelContext(dsChannel, fabsdk.WithUser(user1), fabsdk.WithOrg(org1))
	org2ChannelClientContext := org2sdk.ChannelContext(dsChannel, fabsdk.WithUser(user1), fabsdk.WithOrg(org2))

	// Org1 user connects to 'dschannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	require.NoError(t, err, "Failed to create new channel client for Org1 user: %s", err)

	// Org2 user connects to 'dschannel'
	chClientOrg2User, err := channel.New(org2ChannelClientContext)
	require.NoError(t, err, "Failed to create new channel client for Org1 user: %s", err)

	req := channel.Request{
		ChaincodeID: examplecc,
		Fcn:         "invoke",
		Args:        integration.ExampleCCDefaultQueryArgs(),
	}

	// query org1
	resp, err := chClientOrg1User.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	require.NoError(t, err, "query funds failed")

	foundOrg2Endorser := false
	for _, v := range resp.Responses {
		//check if response endorser is org2 peer and MSP ID 'Org2MSP' is found
		if strings.Contains(string(v.Endorsement.Endorser), "Org2MSP") {
			foundOrg2Endorser = true
			break
		}
	}

	require.True(t, foundOrg2Endorser, "Org2 MSP ID was not in the endorsement")

	//query org2
	resp, err = chClientOrg2User.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	require.NoError(t, err, "query funds failed")

	foundOrg1Endorser := false
	for _, v := range resp.Responses {
		//check if response endorser is org1 peer and MSP ID 'Org1MSP' is found
		if strings.Contains(string(v.Endorsement.Endorser), "Org1MSP") {
			foundOrg1Endorser = true
			break
		}
	}

	require.True(t, foundOrg1Endorser, "Org1 MSP ID was not in the endorsement")
}

// generateChConfigData will generate serialized Channel Config Data for external signing
func generateChConfigData(t *testing.T, dsCtx *dsClientCtx, chConfigPath, user, sigDir string) {
	mspClient, err := mspclient.New(dsCtx.sdk.Context(), mspclient.WithOrg(dsCtx.org))
	require.NoError(t, err, "error creating a new msp management client for %s", dsCtx.org)
	u, err := mspClient.GetSigningIdentity(user)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", dsCtx.org)

	d, err := dsCtx.rsCl.CreateConfigSignatureData(u, chConfigPath)
	require.NoError(t, err, "Failed to fetch Channel config data for signing")

	chCfgName := getBaseChCfgFileName(chConfigPath)

	// create a temporary file and save the channel config data in that file
	f, err := os.Create(path.Join(sigDir, fmt.Sprintf("%s_%s_%s_sbytes.txt", chCfgName, dsCtx.org, user)))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = f.Close()
		require.NoError(t, err, "Failed to close signature file")
	}()

	bufferedWriter := bufio.NewWriter(f)
	_, err = bufferedWriter.Write(d.SigningBytes)
	assert.NoError(t, err, "must be able to write signature of [%s-%s] to buffer", dsCtx.org, user)

	err = bufferedWriter.Flush()
	assert.NoError(t, err, "must be able to flush buffer for signature of [%s-%s] to buffer", dsCtx.org, user)

	// marshal signatureHeader struct for later use
	shf, err := os.Create(path.Join(sigDir, fmt.Sprintf("%s_%s_%s_sHeaderbytes.txt", chCfgName, dsCtx.org, user)))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = shf.Close()
		require.NoError(t, err, "Failed to close signature header file")
	}()

	bufferedWriter = bufio.NewWriter(shf)
	_, err = bufferedWriter.Write(d.SignatureHeaderBytes)
	assert.NoError(t, err, "must be able to write signature header of [%s-%s] to buffer", dsCtx.org, user)

	err = bufferedWriter.Flush()
	assert.NoError(t, err, "must be able to flush buffer for signature of [%s-%s] to buffer", dsCtx.org, user)
}

func generateExternalChConfigSignature(t *testing.T, org, user, chConfigPath, sigDir string) {
	chCfgName := getBaseChCfgFileName(chConfigPath)

	cmd := exec.Command(path.Join("scripts", "generate_signature.sh"), org, user, chCfgName, sigDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	assert.NoError(t, err, "Failed to create external signature for [%s, %s, %s], script error: [%s]", org, user, chCfgName, stderr.String())

	//t.Logf("running generate_signature.sh script output: %s", b)
}

func loadExternalSignature(t *testing.T, org, chConfigPath, user, sigDir string) *common.ConfigSignature {
	chCfgName := getBaseChCfgFileName(chConfigPath)

	fName := path.Join(sigDir, fmt.Sprintf("%s_%s_%s_sbytes.txt.sha256", chCfgName, org, user))
	sig, err := ioutil.ReadFile(fName)
	require.NoError(t, err, "Failed to read signature data with ioutil.ReadFile()")
	//t.Logf("Signature bytes read for %s, %s, %s: '%s'", org, chCfgName, user, sig)

	fName = path.Join(sigDir, fmt.Sprintf("%s_%s_%s_sHeaderbytes.txt", chCfgName, org, user))
	sigHeader, err := ioutil.ReadFile(fName)
	require.NoError(t, err, "Failed to read signature header data")

	isSDKSigning := strings.Contains(chCfgName, "sdk")
	// for signatures signed by OpenSSL, load the private key and convert the signature to LowS as OpenSSL doesn't always return low S as bigInt value.
	// Fabric requires signatures with Low S BigInt value to avoid ECDSA signature malleability.
	// This is required when signing with an external tool like OpenSSL.
	if !isSDKSigning {
		key := loadOrgUserPrivateKey(t, org, user)
		newSig, e := utils.SignatureToLowS(&key.(*ecdsa.PrivateKey).PublicKey, sig)
		require.NoError(t, e, "failed to switch signature to LowS")
		sig = newSig
	}

	cs := &common.ConfigSignature{
		Signature:       sig,
		SignatureHeader: sigHeader,
	}
	return cs
}

//endpointConfigEntity contains endpoint config elements needed by endpointconfig
type endpointConfigEntity struct {
	Channels map[string]fabImpl.ChannelEndpointConfig
}

var clientCfg = `
#
# Schema version of the content. Used by the SDK to apply the corresponding parsing rules.
#
version: 1.0.0

client:

  organization: "org1"

  logging:
    level: "info"

  tlsCerts:
    # [Optional]. Use system certificate pool when connecting to peers, orderers (for negotiating TLS) Default: false
    systemCertPool: true

    # [Optional]. Client key and cert for TLS handshake with peers and orderers
    client:
      key:
        path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.key"
      cert:
        path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.crt"
`

// DynDiscoveryProviderFactory is configured with dynamic (endorser) selection provider
type DynDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateLocalDiscoveryProvider returns a new local dynamic discovery provider
func (f *DynDiscoveryProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return dynamicdiscovery.NewLocalProvider(config), nil
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *DynDiscoveryProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config)
	if err != nil {
		return nil, err
	}
	return &chanProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*dynamicdiscovery.ChannelService),
	}, nil
}

type chanProvider struct {
	fab.ChannelProvider
	services map[string]*dynamicdiscovery.ChannelService
}

func loadOrgUserPrivateKey(t *testing.T, org, user string) interface{} {
	o := strings.ToLower(org)
	pathToKey := path.Join("peerOrganizations", fmt.Sprintf("%s.example.com", o), "users", fmt.Sprintf("%s@%s.example.com", user, o), "msp", "keystore")
	root := integration.GetCryptoConfigPath(pathToKey)
	var parentKey string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "_sk") {
			parentKey = path
		}
		return nil
	})
	require.NoError(t, err, "Failed to fetch _SK file from '%s'", pathToKey)

	key := loadPrivateKey(t, parentKey)

	return key
}

func loadPrivateKey(t *testing.T, path string) interface{} {
	raw, err := ioutil.ReadFile(path)
	require.NoError(t, err, "Failed to read PK @ '%s'", path)

	key, err := utils.PEMtoPrivateKey(raw, []byte(""))
	require.NoError(t, err, "Failed to convert PEM data to PK")

	return key
}

func isOpensslAvailable(t *testing.T) bool {
	cmd := exec.Command(path.Join(string(os.PathSeparator), "bin", "sh"), "-c", "command -v openssl")
	if err := cmd.Run(); err != nil {
		t.Logf("Checking if openssl command is available failed (command -v openssl) [error: %s]. Make sure openssl is installed. Skipping External Channel Config Signing with openssl tests.", err)
		return false
	}
	return true
}
