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

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

var logger = logging.NewLogger("fabsdk/client")

const (
	ccDataProviderSCC      = "lscc"
	ccDataProviderfunction = "getccdata"
)

type peerCreator interface {
	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error)
}

// CCPolicyProvider retrieves policy for the given chaincode ID
type CCPolicyProvider interface {
	GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error)
}

// NewCCPolicyProvider creates new chaincode policy data provider
func newCCPolicyProvider(providers api.Providers, channelID string, userName string, orgName string) (CCPolicyProvider, error) {
	if providers == nil || channelID == "" || userName == "" || orgName == "" {
		return nil, errors.New("Must provide providers, channel ID, user name and organisation for cc policy provider")
	}

	// TODO: Add option to use anchor peers instead of config
	targetPeers, err := providers.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
	}

	//Get identity
	mgr, ok := providers.IdentityManager(orgName)
	if !ok {
		return nil, errors.New("invalid options to create identity, invalid org name")
	}

	identity, err := mgr.GetUser(userName)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create identity for ccl policy provider")
	}

	cpp := ccPolicyProvider{
		config:      providers.Config(),
		providers:   providers,
		channelID:   channelID,
		identity:    identity,
		targetPeers: targetPeers,
		ccDataMap:   make(map[string]*ccprovider.ChaincodeData),
		provider:    providers.InfraProvider(),
	}

	return &cpp, nil
}

type ccPolicyProvider struct {
	config      core.Config
	providers   context.Providers
	channelID   string
	identity    msp.Identity
	targetPeers []core.ChannelPeer
	ccDataMap   map[string]*ccprovider.ChaincodeData // TODO: Add expiry and configurable timeout for map entries
	mutex       sync.RWMutex
	provider    peerCreator
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

	//prepare channel context
	channelContext := dp.getChannelContext()

	//get channel client
	client, err := channel.New(channelContext)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to create channel client")
	}

	for _, p := range dp.targetPeers {

		peer, err := dp.provider.CreatePeerFromConfig(&p.NetworkPeer)
		if err != nil {
			queryErrors = append(queryErrors, err.Error())
			continue
		}

		// Send query to channel peer
		request := channel.Request{
			ChaincodeID: ccID,
			Fcn:         ccFcn,
			Args:        ccArgs,
		}

		resp, err := client.Query(request, channel.WithTargets([]fab.Peer{peer}))
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

func (dp *ccPolicyProvider) getChannelContext() context.ChannelProvider {
	//Get Channel Context
	return func() (context.Channel, error) {
		//Get Client Context
		clientProvider := func() (context.Client, error) {
			return &contextImpl.Client{Providers: dp.providers, Identity: dp.identity}, nil
		}

		return contextImpl.NewChannel(clientProvider, dp.channelID)
	}
}
