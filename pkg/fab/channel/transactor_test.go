/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"

	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	mocksConfig "github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

func TestCreateTxnID(t *testing.T) {
	transactor := createTransactor(t)
	createTxnID(t, transactor)
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

func createTxnID(t *testing.T, transactor *Transactor) fab.TransactionHeader {
	txh, err := transactor.CreateTransactionHeader()
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
	configBackends, err := config.FromFile("../../core/config/testdata/config_test.yaml")()
	if err != nil {
		t.Fatal("failed to get config backends")
	}

	//Override orderer URL in endpoint config
	//Create an empty network config
	networkConfig := fab.NetworkConfig{}
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
		t.Fatal("failed to get endpoint config")
	}

	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	ctx.SetEndpointConfig(endpointCfg)
	chConfig := mocks.NewMockChannelCfg("mychannel")
	chConfig.MockOrderers = []string{"example.com"}

	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
	assert.Equal(t, 1, len(o), "expected one orderer from response orderers list")
	assert.Equal(t, sampleOrdererURL, o[0].URL(), "orderer URL override from endpointconfig channels is not working as expected")
}
