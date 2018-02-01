/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"

	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	ccDataProviderSCC      = "lscc"
	ccDataProviderfunction = "getccdata"
)

// CCPolicyProvider retrieves policy for the given chaincode ID
type CCPolicyProvider interface {
	GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error)
}

// NewCCPolicyProvider creates new chaincode policy data provider
func newCCPolicyProvider(sdk *fabsdk.FabricSDK, channelID string, userName string, orgName string) (CCPolicyProvider, error) {
	if channelID == "" || userName == "" || orgName == "" {
		return nil, errors.New("Must provide channel ID, user name and organisation for cc policy provider")
	}

	if sdk == nil {
		return nil, errors.New("Must provide sdk")
	}

	client := sdk.NewClient(fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))

	// TODO: Add option to use anchor peers instead of config
	targetPeers, err := sdk.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	return &ccPolicyProvider{config: sdk.Config(), client: client, channelID: channelID, targetPeers: targetPeers, ccDataMap: make(map[string]*ccprovider.ChaincodeData)}, nil
}

type ccPolicyProvider struct {
	config      apiconfig.Config
	client      *fabsdk.ClientContext
	channelID   string
	targetPeers []apiconfig.ChannelPeer
	ccDataMap   map[string]*ccprovider.ChaincodeData // TODO: Add expiry and configurable timeout for map entries
	mutex       sync.RWMutex
}

func (dp *ccPolicyProvider) GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error) {
	if chaincodeID == "" {
		return nil, errors.New("Must provide chaincode ID")
	}

	key := newResolverKey(dp.channelID, chaincodeID)
	var ccData *ccprovider.ChaincodeData

	dp.mutex.RLock()
	ccData = dp.ccDataMap[chaincodeID]
	dp.mutex.RUnlock()
	if ccData != nil {
		return unmarshalPolicy(ccData.Policy)
	}

	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	response, err := dp.queryChaincode(ccDataProviderSCC, ccDataProviderfunction, [][]byte{[]byte(dp.channelID), []byte(chaincodeID)})
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error querying chaincode data for chaincode [%s] on channel [%s]", chaincodeID, dp.channelID))
	}

	ccData = &ccprovider.ChaincodeData{}
	err = proto.Unmarshal(response, ccData)
	if err != nil {
		return nil, errors.WithMessage(err, "Error unmarshalling chaincode data")
	}

	dp.ccDataMap[key.String()] = ccData

	return unmarshalPolicy(ccData.Policy)
}

func unmarshalPolicy(policy []byte) (*common.SignaturePolicyEnvelope, error) {

	sigPolicyEnv := &common.SignaturePolicyEnvelope{}
	if err := proto.Unmarshal(policy, sigPolicyEnv); err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling SignaturePolicyEnvelope")
	}

	return sigPolicyEnv, nil
}

func (dp *ccPolicyProvider) queryChaincode(ccID string, ccFcn string, ccArgs [][]byte) ([]byte, error) {
	logger.Debugf("queryChaincode channelID:%s", dp.channelID)

	var queryErrors []string
	var response []byte

	channel, err := dp.client.Channel(dp.channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to create channel client")
	}

	for _, p := range dp.targetPeers {

		peer, err := peerImpl.New(dp.config, peerImpl.FromPeerConfig(&p.NetworkPeer))
		if err != nil {
			queryErrors = append(queryErrors, err.Error())
			continue
		}

		// Send query to channel peer
		request := chclient.Request{
			ChaincodeID: ccID,
			Fcn:         ccFcn,
			Args:        ccArgs,
		}

		resp, err := channel.Query(request, chclient.WithProposalProcessor(peer))
		if err != nil {
			queryErrors = append(queryErrors, err.Error())
			continue
		} else {
			// Valid response obtained, stop querying
			response = resp.Payload
			break
		}
	}
	logger.Debugf("queryErrors: %v", queryErrors)

	// If all queries failed, return error
	if len(queryErrors) == len(dp.targetPeers) {
		errMsg := fmt.Sprintf("Error querying peers for channel %s: %s", dp.channelID, strings.Join(queryErrors, "\n"))
		return nil, errors.New(errMsg)
	}

	return response, nil
}

type resolverKey struct {
	channelID    string
	chaincodeIDs []string
	key          string
}

func (k *resolverKey) String() string {
	return k.key
}

func newResolverKey(channelID string, chaincodeIDs ...string) *resolverKey {
	arr := chaincodeIDs[:]
	sort.Strings(arr)

	key := channelID + "-"
	for i, s := range arr {
		key += s
		if i+1 < len(arr) {
			key += ":"
		}
	}
	return &resolverKey{channelID: channelID, chaincodeIDs: arr, key: key}
}
