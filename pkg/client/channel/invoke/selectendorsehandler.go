/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
)

var logger = logging.NewLogger("fabsdk/client")

var lsccFilter = func(ccID string) bool {
	return ccID != "lscc" && ccID != "_lifecycle"
}

// SelectAndEndorseHandler selects endorsers according to the policies of the chaincodes in the provided invocation chain
// and then sends the proposal to those endorsers. The read/write sets from the responses are then checked to see if additional
// chaincodes were invoked that were not in the original invocation chain. If so, a new endorser set is computed with the
// additional chaincodes and (if necessary) endorsements are requested from those additional endorsers.
type SelectAndEndorseHandler struct {
	*EndorsementHandler
	next Handler
}

// NewSelectAndEndorseHandler returns a new SelectAndEndorseHandler
func NewSelectAndEndorseHandler(next ...Handler) Handler {
	return &SelectAndEndorseHandler{
		EndorsementHandler: NewEndorsementHandler(),
		next:               getNext(next),
	}
}

// Handle selects endorsers and sends proposals to the endorsers
func (e *SelectAndEndorseHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {
	var ccCalls []*fab.ChaincodeCall
	targets := requestContext.Opts.Targets
	if len(targets) == 0 {
		var err error
		ccCalls, requestContext.Opts.Targets, err = getEndorsers(requestContext, clientContext)
		if err != nil {
			requestContext.Error = err
			return
		}
	}

	e.EndorsementHandler.Handle(requestContext, clientContext)

	if requestContext.Error != nil {
		return
	}

	if len(targets) == 0 && len(requestContext.Response.Responses) > 0 {
		additionalEndorsers, err := getAdditionalEndorsers(requestContext, clientContext, ccCalls)
		if err != nil {
			// Log a warning. No need to fail the endorsement. Use the responses collected so far,
			// which may be sufficient to satisfy the chaincode policy.
			logger.Warnf("error getting additional endorsers: %s", err)
		} else {
			if len(additionalEndorsers) > 0 {
				requestContext.Opts.Targets = additionalEndorsers
				logger.Debugf("...getting additional endorsements from %d target(s)", len(additionalEndorsers))
				additionalResponses, err := clientContext.Transactor.SendTransactionProposal(requestContext.Response.Proposal, peer.PeersToTxnProcessors(additionalEndorsers))
				if err != nil {
					requestContext.Error = errors.WithMessage(err, "error sending transaction proposal")
					return
				}

				// Add the new endorsements to the list of responses
				requestContext.Response.Responses = append(requestContext.Response.Responses, additionalResponses...)
			} else {
				logger.Debugf("...no additional endorsements are required.")
			}
		}
	}

	if e.next != nil {
		e.next.Handle(requestContext, clientContext)
	}
}

//NewChainedCCFilter returns a chaincode filter that chains
//multiple filters together. False is returned if at least one
//of the filters in the chain returns false.
func NewChainedCCFilter(filters ...CCFilter) CCFilter {
	return func(ccID string) bool {
		for _, filter := range filters {
			if !filter(ccID) {
				return false
			}
		}
		return true
	}
}

func getEndorsers(requestContext *RequestContext, clientContext *ClientContext, opts ...options.Opt) ([]*fab.ChaincodeCall, []fab.Peer, error) {
	var selectionOpts []options.Opt
	selectionOpts = append(selectionOpts, opts...)
	if requestContext.SelectionFilter != nil {
		selectionOpts = append(selectionOpts, selectopts.WithPeerFilter(requestContext.SelectionFilter))
	}
	if requestContext.PeerSorter != nil {
		selectionOpts = append(selectionOpts, selectopts.WithPeerSorter(requestContext.PeerSorter))
	}

	ccCalls := newInvocationChain(requestContext)
	peers, err := clientContext.Selection.GetEndorsersForChaincode(newInvocationChain(requestContext), selectionOpts...)
	return ccCalls, peers, err
}

func getAdditionalEndorsers(requestContext *RequestContext, clientContext *ClientContext, invocationChain []*fab.ChaincodeCall) ([]fab.Peer, error) {
	invocationChainFromResponse, err := getInvocationChainFromResponse(requestContext.Response.Responses[0])
	if err != nil {
		return nil, errors.WithMessage(err, "error getting invocation chain from proposal response")
	}

	invocationChain, foundAdditional := mergeInvocationChains(invocationChain, invocationChainFromResponse, getCCFilter(requestContext))
	if !foundAdditional {
		return nil, nil
	}

	requestContext.Request.InvocationChain = invocationChain

	logger.Debugf("Found additional chaincodes/collections. Checking if additional endorsements are required...")

	// If using Fabric selection then disable retries. We don't want to keep retrying if the endorsement query returns an error.
	// Also, add a priority selector that gives priority to peers from which we already have endorsements. This way, we don't
	// unnecessarily get endorsements from other orgs.
	_, endorsers, err := getEndorsers(
		requestContext, clientContext,
		selectopts.WithRetryOpts(retry.Opts{}),
		selectopts.WithPrioritySelector(prioritizePeers(requestContext.Opts.Targets)))
	if err != nil {
		return nil, errors.WithMessage(err, "error getting additional endorsers")
	}

	var additionalEndorsers []fab.Peer
	for _, endorser := range endorsers {
		if !containsMSP(requestContext.Opts.Targets, endorser.MSPID()) {
			logger.Debugf("... will ask for additional endorsement from [%s] in order to satisfy the chaincode policy", endorser.URL())
			additionalEndorsers = append(additionalEndorsers, endorser)
		}
	}

	return additionalEndorsers, nil
}

func getCCFilter(requestContext *RequestContext) CCFilter {
	if requestContext.Opts.CCFilter != nil {
		return NewChainedCCFilter(lsccFilter, requestContext.Opts.CCFilter)
	}
	return lsccFilter
}

func containsMSP(peers []fab.Peer, mspID string) bool {
	for _, p := range peers {
		if p.MSPID() == mspID {
			return true
		}
	}
	return false
}

func getInvocationChainFromResponse(response *fab.TransactionProposalResponse) ([]*fab.ChaincodeCall, error) {
	rwSets, err := getRWSetsFromProposalResponse(response.ProposalResponse)
	if err != nil {
		return nil, err
	}

	invocationChain := make([]*fab.ChaincodeCall, len(rwSets))
	for i, rwSet := range rwSets {
		collections := make([]string, len(rwSet.CollHashedRwSets))
		for j, collRWSet := range rwSet.CollHashedRwSets {
			collections[j] = collRWSet.CollectionName
		}
		logger.Debugf("Found chaincode in RWSet [%s], Collections %v", rwSet.NameSpace, collections)
		invocationChain[i] = &fab.ChaincodeCall{ID: rwSet.NameSpace, Collections: collections}
	}

	return invocationChain, nil
}

func getRWSetsFromProposalResponse(response *pb.ProposalResponse) ([]*rwsetutil.NsRwSet, error) {
	if response == nil {
		return nil, nil
	}

	prp := &pb.ProposalResponsePayload{}
	err := proto.Unmarshal(response.Payload, prp)
	if err != nil {
		return nil, err
	}

	chaincodeAction := &pb.ChaincodeAction{}
	err = proto.Unmarshal(prp.Extension, chaincodeAction)
	if err != nil {
		return nil, err
	}

	if len(chaincodeAction.Results) == 0 {
		return nil, nil
	}

	txRWSet := &rwsetutil.TxRwSet{}
	if err := txRWSet.FromProtoBytes(chaincodeAction.Results); err != nil {
		return nil, err
	}

	return txRWSet.NsRwSets, nil
}

func mergeInvocationChains(invocChain []*fab.ChaincodeCall, respInvocChain []*fab.ChaincodeCall, filter CCFilter) ([]*fab.ChaincodeCall, bool) {
	var mergedInvocChain []*fab.ChaincodeCall
	var changed bool
	for _, respCCCall := range respInvocChain {
		if !filter(respCCCall.ID) {
			logger.Debugf("Ignoring chaincode [%s] in the RW set since it was filtered out", respCCCall.ID)
			continue
		}
		mergedCCCall, merged := mergeCCCall(invocChain, respCCCall)
		if merged {
			changed = true
		}
		mergedInvocChain = append(mergedInvocChain, mergedCCCall)
	}
	return mergedInvocChain, changed
}

// mergeCCCall checks if the provided invocation chain contains the given Chaincode Call.
// - If the invocation chain does not contain the chaincode call then return (respCCCall,true)
// - If the invocation chain contains the chaincode call but the collection sets are different, then return (mergedCCCall,true)
// - If the invocation chain contains the chaincode call and the collection sets are the same, then return (respCCCall,false)
func mergeCCCall(invocChain []*fab.ChaincodeCall, respCCCall *fab.ChaincodeCall) (*fab.ChaincodeCall, bool) {
	ccCall, ok := getCCCall(invocChain, respCCCall.ID)
	if ok {
		logger.Debugf("Already have chaincode [%s]. Checking to see if any private data collections were detected in the proposal response", respCCCall.ID)
		c, merged := merge(ccCall, respCCCall)
		if merged {
			logger.Debugf("Modifying chaincode call for chaincode [%s] since additional private data collections were detected in the RW set", respCCCall.ID)
		} else {
			logger.Debugf("No additional private data collections were detected for chaincode [%s]", respCCCall.ID)
		}
		return c, merged
	}

	logger.Debugf("Detected chaincode [%s] in the RW set of the proposal response that was not part of the original invocation chain", respCCCall.ID)
	return respCCCall, true
}

// getCC returns the ChaincodeCall from the invocation chain that matches the chaincode ID or
// returns nil if the ChaincodeCall is not found.
func getCCCall(invocChain []*fab.ChaincodeCall, ccID string) (*fab.ChaincodeCall, bool) {
	for _, ccCall := range invocChain {
		if ccCall.ID == ccID {
			return ccCall, true
		}
	}
	return nil, false
}

// merge merges the collections from c1 and c2 and returns the resulting ChaincodeCall.
// true is returned if a merge was necessary; false is returned if the two ChaincodeCalls were the same.
func merge(c1 *fab.ChaincodeCall, c2 *fab.ChaincodeCall) (*fab.ChaincodeCall, bool) {
	c := &fab.ChaincodeCall{ID: c1.ID, Collections: c1.Collections}
	merged := false
	for _, coll := range c2.Collections {
		if !contains(c.Collections, coll) {
			c.Collections = append(c.Collections, coll)
			merged = true
		}
	}
	return c, merged
}

func contains(values []string, value string) bool {
	for _, val := range values {
		if val == value {
			return true
		}
	}
	return false
}

// prioritizePeers is a priority selector that gives priority to the peers that are in the given set
func prioritizePeers(peers []fab.Peer) selectopts.PrioritySelector {
	return func(peer1, peer2 fab.Peer) int {
		hasPeer1 := containsPeer(peers, peer1)
		hasPeer2 := containsPeer(peers, peer2)

		if hasPeer1 && hasPeer2 {
			return 0
		}
		if hasPeer1 {
			return 1
		}
		if hasPeer2 {
			return -1
		}
		return 0
	}
}

func containsPeer(peers []fab.Peer, peer fab.Peer) bool {
	for _, p := range peers {
		if p.URL() == peer.URL() {
			return true
		}
	}
	return false
}
