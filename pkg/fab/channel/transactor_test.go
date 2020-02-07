/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	mocksConfig "github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

func TestCreateTxnID(t *testing.T) {
	transactor := createTransactor(t)

	txh := createTxnID(t, transactor)
	assert.NotEmpty(t, txh.Nonce())
	assert.NotEmpty(t, txh.Creator())
	assert.NotEmpty(t, txh.TransactionID())

	creator := []byte("creator")
	nonce := []byte("12345")

	txh = createTxnID(t, transactor, fab.WithCreator(creator), fab.WithNonce(nonce))
	assert.Equal(t, nonce, txh.Nonce())
	assert.Equal(t, creator, txh.Creator())
	assert.NotEmpty(t, txh.TransactionID())
}

func TestTransactionProposal(t *testing.T) {
	transactor := createTransactor(t)
	tp := createTransactionProposal(t, transactor)
	createTransactionProposalResponse(t, transactor, tp)
}

func TestTransaction(t *testing.T) {
	transactor := createTransactor(t)
	tp := createTransactionProposal(t, transactor)
	tpr := createTransactionProposalResponse(t, transactor, tp)

	request := fab.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: tpr,
	}
	tx, err := txn.New(request)
	assert.Nil(t, err)

	_, err = transactor.SendTransaction(tx)
	assert.Nil(t, err)
}

func TestTransactionBadStatus(t *testing.T) {
	transactor := createTransactor(t)
	tp := createTransactionProposal(t, transactor)
	tpr := createTransactionProposalResponseBadStatus(t, transactor, tp)

	request := fab.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: tpr,
	}
	_, err := txn.New(request)
	assert.NotNil(t, err)
}

func createTransactor(t *testing.T) *Transactor {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	orderer := mocks.NewMockOrderer("", nil)
	chConfig := mocks.NewMockChannelCfg("testChannel")
	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()
	transactor, err := NewTransactor(reqCtx, chConfig)
	transactor.orderers = []fab.Orderer{orderer}
	assert.Nil(t, err)

	return transactor
}

func createTxnID(t *testing.T, transactor *Transactor, opts ...fab.TxnHeaderOpt) fab.TransactionHeader {
	txh, err := transactor.CreateTransactionHeader(opts...)
	assert.Nil(t, err, "creation of transaction ID failed")

	return txh
}

func createTransactionProposal(t *testing.T, transactor *Transactor) *fab.TransactionProposal {
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "example",
		Fcn:         "fcn",
	}

	txh := createTxnID(t, transactor)
	tp, err := txn.CreateChaincodeInvokeProposal(txh, request)
	assert.Nil(t, err)

	assert.NotEmpty(t, tp.TxnID)

	return tp
}

func createTransactionProposalResponse(t *testing.T, transactor fab.Transactor, tp *fab.TransactionProposal) []*fab.TransactionProposalResponse {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, Status: 200}
	tpr, err := transactor.SendTransactionProposal(tp, []fab.ProposalProcessor{&peer})
	assert.Nil(t, err)

	return tpr
}

func createTransactionProposalResponseBadStatus(t *testing.T, transactor fab.Transactor, tp *fab.TransactionProposal) []*fab.TransactionProposalResponse {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, Status: 500}
	tpr, err := transactor.SendTransactionProposal(tp, []fab.ProposalProcessor{&peer})
	assert.Nil(t, err)

	return tpr
}

// TestOrderersFromChannelCfg uses an orderer that exists in the configuration.
func TestOrderersFromChannelCfg(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	chConfig := mocks.NewMockChannelCfg("testChannel")
	chConfig.MockOrderers = []string{"example.com"}

	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
}

// TestOrderersFromChannel - tests scenario where err should not be returned if channel config is not found
//instead, empty orderers list should be returned
func TestOrderersFromChannel(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)

	o, err := orderersFromChannel(ctx, "invalid-channel-id")
	assert.Nil(t, err)
	assert.NotNil(t, o)
	assert.Zero(t, len(o))
}

// TestOrderersFromChannelCfg uses an orderer that does not exist in the configuration.
func TestOrderersFromChannelCfgBadTLS(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	chConfig := mocks.NewMockChannelCfg("testChannel")
	chConfig.MockOrderers = []string{"doesnotexist.com"}

	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
}

// TestOrderersURLOverride tests orderer URL override from endpoint channels config
func TestOrderersURLOverride(t *testing.T) {
	sampleOrdererURL := "orderer.example.com.sample.url:100090"

	//Create endpoint config
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", "config_test.yaml")
	configBackends, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("failed to get config backends")
	}

	//Override orderer URL in endpoint config
	//Create an empty network config
	networkConfig := endpointConfigEntity{}
	err = lookup.New(configBackends...).UnmarshalKey("orderers", &networkConfig.Orderers)
	if err != nil {
		t.Fatal("failed to unmarshal orderer")
	}

	orderer := networkConfig.Orderers["orderer.example.com"]
	orderer.URL = sampleOrdererURL
	networkConfig.Orderers["orderer.example.com"] = orderer

	backendMap := make(map[string]interface{})
	backendMap["orderers"] = networkConfig.Orderers
	backends := append([]core.ConfigBackend{}, &mocksConfig.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, configBackends...)
	endpointCfg, err := fabImpl.ConfigFromBackend(backends...)
	if err != nil {
		t.Fatal("failed to get endpoint config", err)
	}

	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	ctx.SetEndpointConfig(endpointCfg)
	// create a mock channel config for mychannel with this orderer created above
	chConfig := mocks.NewMockChannelCfg("mychannel")
	chConfig.MockOrderers = []string{"example.com"}

	// now test orderersFromChannelCfg with above channel config (chConfig) and sdk config passed in as ctx
	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 1, len(o), "expected one orderer from response orderers list")
	assert.Equal(t, sampleOrdererURL, o[0].URL(), "orderer URL override from endpointconfig channels is not working as expected")
}

func TestExcludedOrdrerer(t *testing.T) {
	sampleOrdererURL := "orderer.example.com.sample.url:100090"

	//Create endpoint mockEndpoingCfg
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", "config_test.yaml")
	configBackends, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("failed to get mockEndpoingCfg backends")
	}

	networkConfig := endpointConfigEntity{}
	err = lookup.New(configBackends...).UnmarshalKey("orderers", &networkConfig.Orderers)
	if err != nil {
		t.Fatal("failed to unmarshal orderer")
	}

	orderer := networkConfig.Orderers["orderer.example.com"]
	orderer.URL = sampleOrdererURL
	networkConfig.Orderers["orderer.example.com"] = orderer

	// create a mock channel mockEndpoingCfg for mychannel with this orderer created above
	chConfig := mocks.NewMockChannelCfg("mychannel")
	chConfig.MockOrderers = []string{"example.com"}

	backendMap := make(map[string]interface{})
	backendMap["orderers"] = networkConfig.Orderers

	backends := append([]core.ConfigBackend{}, &mocksConfig.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, configBackends...)

	// now try to add a second orderer to the configs
	// 1. update channel mockEndpoingCfg with this new orderer
	chConfig.MockOrderers = append(chConfig.MockOrderers, "example2.com")
	// 2. update sdk configs as well
	sampleOrderer2URL := "orderer.example2.com:9999"
	networkConfig.Orderers["orderer.example2.com"] = fabImpl.OrdererConfig{
		URL:         sampleOrderer2URL,
		TLSCACerts:  networkConfig.Orderers["orderer.example.com"].TLSCACerts,  // for testing only, adding dummy cert
		GRPCOptions: networkConfig.Orderers["orderer.example.com"].GRPCOptions, // for testing only, adding dummy cert
	}
	backendMap["orderers"] = networkConfig.Orderers

	backendMap["channels"] = map[string]interface{}{
		"mychannel": map[string]interface{}{
			"orderers": []string{"orderer.example.com", "orderer.example2.com"},
		},
	}

	backends = append([]core.ConfigBackend{}, &mocksConfig.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, configBackends...)
	endpointCfg, err := fabImpl.ConfigFromBackend(backends...)
	if err != nil {
		t.Fatal("failed to get endpoint mockEndpoingCfg", err)
	}

	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	ctx.SetEndpointConfig(endpointCfg)
	// 3. now test orderersFromChannelCfg with updated chConfig and sdk mockEndpoingCfg (ctx)
	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 2, len(o), "expected 2 orderers from response orderers list")
	assert.Equal(t, sampleOrdererURL, o[0].URL(),
		"orderer URL override from endpointconfig channels is not working as expected")
	assert.Equal(t, sampleOrderer2URL, o[1].URL(),
		"orderer URL override from endpointconfig channels is not working as expected")

	sampleOrdererExcludedURL := "orderer.excluded.example3.com:8888"
	// finally add a blacklisted orderer and ensure it's not returned in the configs

	// first make sure it's added in the channel config
	chConfig.MockOrderers = append(chConfig.MockOrderers, "example3.com")

	// and add it's added in the networkConfig of the SDK
	networkConfig.Orderers[sampleOrdererExcludedURL] = fabImpl.OrdererConfig{
		URL:         sampleOrdererExcludedURL,
		TLSCACerts:  networkConfig.Orderers["orderer.example.com"].TLSCACerts,  // for testing only, adding dummy cert
		GRPCOptions: networkConfig.Orderers["orderer.example.com"].GRPCOptions, // for testing only, adding dummy cert
	}

	// create mock EncdpointConfig to control returned values
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockEndpoingCfg := mockfab.NewMockEndpointConfig(mockCtrl)

	var orderersCfgs []fab.OrdererConfig
	for _, v := range networkConfig.Orderers {
		orderersCfgs = append(orderersCfgs, fab.OrdererConfig{
			URL: v.URL,
			GRPCOptions: v.GRPCOptions,
		})
	}

	mockEndpoingCfg.EXPECT().ChannelConfig("mychannel").Return(&fab.ChannelEndpointConfig{
		Orderers: []string{}})  // empty channel.Orderers SDK config, to force fetching from orderers SDK config
	mockEndpoingCfg.EXPECT().OrdererConfig("example.com").Return(&orderersCfgs[0], true, false)
	mockEndpoingCfg.EXPECT().OrdererConfig("example2.com").Return(&orderersCfgs[1], true, false)
	mockEndpoingCfg.EXPECT().OrdererConfig("example3.com").Return(nil, false, true) // true means ignored
	mockEndpoingCfg.EXPECT().OrderersConfig().Return(orderersCfgs)

	ctx.SetEndpointConfig(mockEndpoingCfg)

	// example3.com is marked as ignored in mockEndpointCfg above (this is equivalent to field: ignoreEndpoint:true in
	// EntityMatchers) in SDK configs while it is added in the channel chConfig
	o, err = orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 2, len(o), "expected 2 orderers from response orderers list")

	// now try with example2.com not found, to be populated from chConfig
	mockEndpoingCfg.EXPECT().ChannelConfig("mychannel").Return(&fab.ChannelEndpointConfig{
		Orderers: []string{}}) // empty channel.Orderers SDK config, to force fetching from orderers SDK config
	mockEndpoingCfg.EXPECT().OrdererConfig("example.com").Return(&orderersCfgs[0], true, false)
	mockEndpoingCfg.EXPECT().OrdererConfig("example2.com").Return(nil, false, false)
	mockEndpoingCfg.EXPECT().OrdererConfig("example3.com").Return(nil, false, true) // true means ignored
	mockEndpoingCfg.EXPECT().OrderersConfig().Return(orderersCfgs)

	ctx.SetEndpointConfig(mockEndpoingCfg)

	o, err = orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 2, len(o), "expected 2 orderers from response orderers list")


	// now retry the same previous two tests with channelConfig returning list of orderers
	mockEndpoingCfg.EXPECT().ChannelConfig("mychannel").Return(&fab.ChannelEndpointConfig{
		Orderers: chConfig.MockOrderers}) // read orderers from channel.Orderers SDK config
	mockEndpoingCfg.EXPECT().OrdererConfig("example.com").Return(&orderersCfgs[0], true, false)
	mockEndpoingCfg.EXPECT().OrdererConfig("example2.com").Return(&orderersCfgs[1], true, false)
	mockEndpoingCfg.EXPECT().OrdererConfig("example3.com").Return(nil, false, true) // true means ignored

	ctx.SetEndpointConfig(mockEndpoingCfg)

	// example3.com is marked as ignored in mockEndpointCfg above (this is equivalent to field: ignoreEndpoint:true in
	// EntityMatchers) in SDK configs while it is added in the channel chConfig
	o, err = orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 2, len(o), "expected 2 orderers from response orderers list")

	// now try with example2.com not found, to be populated from chConfig
	mockEndpoingCfg.EXPECT().ChannelConfig("mychannel").Return(&fab.ChannelEndpointConfig{
		Orderers: chConfig.MockOrderers})  // read orderers from channel.Orderers SDK config
	mockEndpoingCfg.EXPECT().OrdererConfig("example.com").Return(&orderersCfgs[0], true, false) // found
	mockEndpoingCfg.EXPECT().OrdererConfig("example2.com").Return(nil, false, false) // not found
	mockEndpoingCfg.EXPECT().OrdererConfig("example3.com").Return(nil, false, true) // excluded

	ctx.SetEndpointConfig(mockEndpoingCfg)

	o, err = orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 1, len(o),
		"expected 1 orderer from response orderers list since 1 orderer is not found " +
		"and another is excluded")
}

//endpointConfigEntity contains endpoint config elements needed by endpointconfig
type endpointConfigEntity struct {
	Channels      map[string]fab.ChannelEndpointConfig
	Organizations map[string]fabImpl.OrganizationConfig
	Orderers      map[string]fabImpl.OrdererConfig
	Peers         map[string]fabImpl.PeerConfig
}
