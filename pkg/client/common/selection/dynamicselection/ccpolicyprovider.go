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
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
)

const loggerModule = "fabsdk/client"

var logger = logging.NewLogger(loggerModule)

const (
	ccDataProviderSCC      = "lscc"
	ccDataProviderfunction = "getccdata"
)

// CCPolicyProvider retrieves policy for the given chaincode ID
type CCPolicyProvider interface {
	GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error)
}

// NewCCPolicyProvider creates new chaincode policy data provider
func newCCPolicyProvider(ctx context.Client, discovery fab.DiscoveryService, channelID string) (CCPolicyProvider, error) {
	if channelID == "" {
		return nil, errors.New("Must provide channel ID for cc policy provider")
	}

	cpp := ccPolicyProvider{
		context:   ctx,
		channelID: channelID,
		discovery: discovery,
		ccDataMap: make(map[string]*ccprovider.ChaincodeData),
	}

	return &cpp, nil
}

type ccPolicyProvider struct {
	context   context.Client
	channelID string
	discovery fab.DiscoveryService
	ccDataMap map[string]*ccprovider.ChaincodeData // TODO: Add expiry and configurable timeout for map entries
	mutex     sync.RWMutex
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
		return nil, errors.WithMessagef(err, "error querying chaincode data for chaincode [%s] on channel [%s]", chaincodeID, dp.channelID)
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

	//prepare channel context
	channelContext := dp.getChannelContext()

	//get channel client
	client, err := channel.New(channelContext)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to create channel client")
	}

	targetPeers, err := dp.discovery.GetPeers()
	if err != nil {
		return nil, status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), err.Error(), nil)
	}

	for _, peer := range targetPeers {

		// Send query to channel peer
		request := channel.Request{
			ChaincodeID: ccID,
			Fcn:         ccFcn,
			Args:        ccArgs,
		}

		resp, err := client.Query(request, channel.WithTargets(peer))
		if err != nil {
			logger.Debugf("query peer '%s' returned error for ccID %s, Fcn %s: %s", peer.URL(), ccID, ccFcn, err)
			queryErrors = append(queryErrors, err.Error())
			continue
		} else {
			// Valid response obtained, stop querying
			response = resp.Payload
			break
		}
	}
	logger.Debugf("queryErrors found %d error(s) from %d peers: %+v", len(queryErrors), len(targetPeers), queryErrors)

	// If all queries failed, return error
	if len(queryErrors) == len(targetPeers) {
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

func (dp *ccPolicyProvider) getChannelContext() context.ChannelProvider {
	//Get Channel Context
	return func() (context.Channel, error) {
		//Get Client Context
		clientProvider := func() (context.Client, error) {
			return dp.context, nil
		}

		return contextImpl.NewChannel(clientProvider, dp.channelID)
	}
}
