/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"
	"fmt"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// TestTransient ...
func TestTransient(t *testing.T) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	fcn := "invoke"
	transientData := "Transient data test..."

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	_, cancel, transactor, err := getTransactor(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	if err != nil {
		t.Fatalf("Failed to get channel transactor: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets[:1])
	require.Nil(t, err, "creating peers failed")

	transactionProposalResponse, _, err := createAndSendTransactionProposal(transactor, chaincodeID, fcn, integration.ExampleCCDefaultTxArgs(), peers, transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %s", err)
	}
	strResponse := string(transactionProposalResponse[0].ProposalResponse.GetResponse().Payload)
	//validate transient data exists in proposal
	if len(strResponse) == 0 {
		t.Fatalf("Transient data does not exist: expected %s", transientData)
	}

	//verify transient data content
	if strResponse != transientData {
		t.Fatalf("Expected '%s' in transient data field. Received '%s' ", transientData, strResponse)
	}
	//transient data null
	transientDataMap["result"] = []byte{}
	transactionProposalResponse, _, err = createAndSendTransactionProposal(transactor, chaincodeID, fcn, integration.ExampleCCDefaultTxArgs(), peers, transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal with empty transient data return an error: %s", err)
	}
	//validate that transient data does not exist in proposal
	strResponse = string(transactionProposalResponse[0].ProposalResponse.GetResponse().Payload)
	if len(strResponse) != 0 {
		t.Fatalf("Transient data validation has failed. An empty transient data was expected but %s was returned", strResponse)
	}

}

// createAndSendTransactionProposal uses transactor to create and send transaction proposal
func createAndSendTransactionProposal(transactor fab.ProposalSender, chainCodeID string,
	fcn string, args [][]byte, targets []fab.ProposalProcessor, transientData map[string][]byte) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, error) {

	propReq := fab.ChaincodeInvokeRequest{
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}

	type invokerResponse struct {
		tpr []*fab.TransactionProposalResponse
		tp  *fab.TransactionProposal
	}

	invokerResp, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			txh, err := transactor.CreateTransactionHeader()
			if err != nil {
				return nil, errors.WithMessage(err, "creating transaction header failed")
			}

			tp, err := txn.CreateChaincodeInvokeProposal(txh, propReq)
			if err != nil {
				return nil, errors.WithMessage(err, "creating transaction proposal failed")
			}

			tpr, err := transactor.SendTransactionProposal(tp, targets)
			if err != nil && strings.Contains(err.Error(), "500") {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("creating transaction proposal failed error: %v", err), nil)
			}

			return &invokerResponse{
				tpr,
				tp,
			}, nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	resp := invokerResp.(*invokerResponse)
	return resp.tpr, resp.tp, err
}

func getTransactor(sdk *fabsdk.FabricSDK, channelID string, user string, orgName string) (reqContext.Context, reqContext.CancelFunc, fab.Transactor, error) {

	clientChannelContextProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	channelContext, err := clientChannelContextProvider()
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "channel service creation failed")
	}
	chService := channelContext.ChannelService()

	chConfig, err := chService.ChannelConfig()
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "channel config retrieval failed")
	}

	reqCtx, cancel := context.NewRequest(channelContext, context.WithTimeoutType(fab.PeerResponse))
	transactor, err := channel.NewTransactor(reqCtx, chConfig)

	return reqCtx, cancel, transactor, err
}

func getProposalProcessors(sdk *fabsdk.FabricSDK, user string, orgName string, targets []string) ([]fab.ProposalProcessor, error) {
	ctxProvider := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	ctx, err := ctxProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "context creation failed")
	}

	var peers []fab.ProposalProcessor
	for _, url := range targets {
		p, err := getPeer(ctx, url)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}

	return peers, nil
}

func getPeer(ctx contextAPI.Client, url string) (fab.Peer, error) {

	peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), url)
	if err != nil {
		return nil, err
	}

	peer, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
	if err != nil {
		return nil, errors.WithMessage(err, "creating peer from config failed")
	}

	return peer, nil
}
