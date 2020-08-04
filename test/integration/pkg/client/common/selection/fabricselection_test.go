// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package selection

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"

	"github.com/stretchr/testify/require"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/fabricselection"
	selectionopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	grpcCodes "google.golang.org/grpc/codes"
)

const (
	peer0Org1 = "peer0.org1.example.com"
	peer1Org1 = "peer1.org1.example.com"
	peer0Org2 = "peer0.org2.example.com"
	peer1Org2 = "peer1.org2.example.com"
)

var (
	hostToURLMap = map[string]string{
		peer0Org1: "peer0.org1.example.com:7051",
		peer1Org1: "peer1.org1.example.com:7151",
		peer0Org2: "peer0.org2.example.com:8051",
		peer1Org2: "peer1.org2.example.com:9051",
	}

	localHostToURLMap = map[string]string{
		peer0Org1: "localhost:7051",
		peer1Org1: "localhost:7151",
		peer0Org2: "localhost:8051",
		peer1Org2: "localhost:9051",
	}
)

var testRetryOpts = retry.Opts{
	Attempts:       10,
	InitialBackoff: 1 * time.Second,
	MaxBackoff:     15 * time.Second,
	BackoffFactor:  2,
	RetryableCodes: map[status.Group][]status.Code{
		status.GRPCTransportStatus: {
			status.Code(grpcCodes.Unavailable),
		},
		status.DiscoveryServerStatus: {
			status.QueryEndorsers,
		},
	},
}

func TestFabricSelection(t *testing.T) {
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithServicePkg(&fabricSelectionProviderFactory{}))
	require.NoError(t, err, "Failed to create new SDK")
	defer sdk.Close()

	orgsContext := setupMultiOrgContext(t, sdk)
	err = integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	selectionService, err := ctx.ChannelService().Selection()
	require.NoError(t, err)

	_, ok := selectionService.(*fabricselection.Service)
	require.True(t, ok)

	t.Run("Policy: Org1 Only", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member')")
			require.NoError(t, err)
		} else {
			err = integration.InstantiateExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "OR('Org1MSP.member')")
			require.NoError(t, err)
		}

		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID)),
			expecting(
				[]string{getURL(peer0Org1)},
				[]string{getURL(peer1Org1)}),
		)
	})

	t.Run("Policy: Org2 Only", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org2MSP.member')")
			require.NoError(t, err)
		} else {
			err = integration.InstantiateExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "OR('Org2MSP.member')")
			require.NoError(t, err)
		}

		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID)),
			expecting(
				[]string{getURL(peer0Org2)},
				[]string{getURL(peer1Org2)}),
		)
	})

	t.Run("Policy: Org1 or Org2", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
			require.NoError(t, err)
		} else {
			err = integration.InstantiateExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
			require.NoError(t, err)
		}

		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID)),
			expecting(
				[]string{getURL(peer0Org1)},
				[]string{getURL(peer1Org1)},
				[]string{getURL(peer0Org2)},
				[]string{getURL(peer1Org2)}),
		)
	})

	t.Run("Policy: Org1 and Org2", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "AND('Org1MSP.member','Org2MSP.member')")
			require.NoError(t, err)
		} else {
			err = integration.InstantiateExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "AND('Org1MSP.member','Org2MSP.member')")
			require.NoError(t, err)
		}

		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID)),
			expecting(
				[]string{getURL(peer0Org1), getURL(peer0Org2)},
				[]string{getURL(peer0Org1), getURL(peer1Org2)},
				[]string{getURL(peer1Org1), getURL(peer0Org2)},
				[]string{getURL(peer1Org1), getURL(peer1Org2)}),
		)

		// With peer filter
		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID)),
			expecting(
				[]string{getURL(peer1Org1), getURL(peer0Org2)},
				[]string{getURL(peer1Org1), getURL(peer1Org2)}),
			selectionopts.WithPeerFilter(func(peer fab.Peer) bool {
				return peer.URL() != getURL(peer0Org1)
			}),
		)
	})

	// Chaincode to Chaincode
	t.Run("Policy: CC1(Org1 Only) to CC2(Org2 Only)", func(t *testing.T) {
		ccID1 := integration.GenerateExampleID(true)
		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID1)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID1, "OR('Org1MSP.member')")
			require.NoError(t, err)
		} else {
			err = integration.InstantiateExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID1, "OR('Org1MSP.member')")
			require.NoError(t, err)
		}

		ccID2 := integration.GenerateExampleID(true)
		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID2)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID2, "OR('Org2MSP.member')")
			require.NoError(t, err)
		} else {
			err = integration.InstantiateExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID2, "OR('Org2MSP.member')")
			require.NoError(t, err)
		}

		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID1), newCCCall(ccID2)),
			expecting(
				[]string{getURL(peer0Org1), getURL(peer0Org2)},
				[]string{getURL(peer1Org1), getURL(peer0Org2)},
				[]string{getURL(peer0Org1), getURL(peer1Org2)},
				[]string{getURL(peer1Org1), getURL(peer1Org2)}),
		)
	})

	t.Run("Policy: Org1 or Org2; ColPolicy: Org1 only", func(t *testing.T) {
		coll1 := "collection1"
		ccID := integration.GenerateExampleID(true)
		collConfig, err := newCollectionConfig(coll1, "OR('Org1MSP.member')", 0, 2, 1000)
		require.NoError(t, err)

		if metadata.CCMode == "lscc" {
			err = integration.InstallExampleChaincode(orgsContext, ccID)
			require.NoError(t, err)
			err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
			require.NoError(t, err)
		} else {
			err = integration.InstantiatePvtExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
			require.NoError(t, err)
		}

		testEndorsers(
			t, selectionService,
			chaincodes(newCCCall(ccID, coll1)),
			expecting(
				[]string{getURL(peer0Org1)},
				[]string{getURL(peer1Org1)}),
		)
	})
}

func testEndorsers(t *testing.T, selectionService fab.SelectionService, chaincodes []*fab.ChaincodeCall, expectedEndorserGroups [][]string, opts ...options.Opt) {
	// Get endorsers a few times, since each time a different set may be returned
	for i := 0; i < 5; i++ {
		endorsers, err := selectionService.GetEndorsersForChaincode(chaincodes, opts...)
		require.NoError(t, err, "error getting endorsers")
		checkEndorsers(t, endorsers, expectedEndorserGroups)
	}
}

func checkEndorsers(t *testing.T, endorsers []fab.Peer, expectedGroups [][]string) {
	for _, group := range expectedGroups {
		if containsAll(t, endorsers, group) {
			t.Logf("Found matching endorser group: %#v", group)
			return
		}
	}
	for _, endorser := range endorsers {
		t.Logf("endorser URL: %s", endorser.URL())
	}

	t.Fatalf("Unexpected endorser group: %#v - Expecting one of: %#v", endorsers, expectedGroups)
}

func containsAll(t *testing.T, endorsers []fab.Peer, expectedEndorserGroup []string) bool {
	if len(endorsers) != len(expectedEndorserGroup) {
		return false
	}

	for _, endorser := range endorsers {
		t.Logf("Checking endpoint: %s ...", endorser)
		if !contains(expectedEndorserGroup, endorser.URL()) {
			return false
		}
	}
	return true
}

func newCCCall(ccID string, collections ...string) *fab.ChaincodeCall {
	return &fab.ChaincodeCall{
		ID:          ccID,
		Collections: collections,
	}
}

func chaincodes(ccCalls ...*fab.ChaincodeCall) []*fab.ChaincodeCall {
	return ccCalls
}

type fabricSelectionProviderFactory struct {
	defsvc.ProviderFactory
}

type fabricSelectionChannelProvider struct {
	fab.ChannelProvider
	services map[string]*fabricselection.Service
}

type fabricSelectionChannelService struct {
	fab.ChannelService
	selection fab.SelectionService
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *fabricSelectionProviderFactory) CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config, opts...)
	if err != nil {
		return nil, err
	}
	return &fabricSelectionChannelProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*fabricselection.Service),
	}, nil
}

// Close frees resources and caches.
func (cp *fabricSelectionChannelProvider) Close() {
	if c, ok := cp.ChannelProvider.(closable); ok {
		c.Close()
	}
	for _, selection := range cp.services {
		selection.Close()
	}
}

type providerInit interface {
	Initialize(providers contextApi.Providers) error
}

func (cp *fabricSelectionChannelProvider) Initialize(providers contextApi.Providers) error {
	if init, ok := cp.ChannelProvider.(providerInit); ok {
		return init.Initialize(providers)
	}
	return nil
}

// ChannelService creates a ChannelService for an identity
func (cp *fabricSelectionChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	chService, err := cp.ChannelProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, err
	}

	discovery, err := chService.Discovery()
	if err != nil {
		return nil, err
	}

	selection, ok := cp.services[channelID]
	if !ok {
		selection, err = fabricselection.New(ctx, channelID, discovery, selectionopts.WithRetryOpts(testRetryOpts))
		if err != nil {
			return nil, err
		}
		cp.services[channelID] = selection
	}

	return &fabricSelectionChannelService{
		ChannelService: chService,
		selection:      selection,
	}, nil
}

func (cs *fabricSelectionChannelService) Selection() (fab.SelectionService, error) {
	return cs.selection, nil
}

func expecting(groups ...[]string) [][]string {
	return groups
}

func newCollectionConfig(colName, policy string, reqPeerCount, maxPeerCount int32, blockToLive uint64) (*pb.CollectionConfig, error) {
	p, err := policydsl.FromString(policy)
	if err != nil {
		return nil, err
	}
	cpc := &pb.CollectionPolicyConfig{
		Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
			SignaturePolicy: p,
		},
	}
	return &pb.CollectionConfig{
		Payload: &pb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &pb.StaticCollectionConfig{
				Name:              colName,
				MemberOrgsPolicy:  cpc,
				RequiredPeerCount: reqPeerCount,
				MaximumPeerCount:  maxPeerCount,
				BlockToLive:       blockToLive,
			},
		},
	}, nil
}

func getURL(host string) string {
	if integration.IsLocal() {
		return localHostToURLMap[host]
	}
	return hostToURLMap[host]
}

type closable interface {
	Close()
}

func contains(list []string, value string) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}
