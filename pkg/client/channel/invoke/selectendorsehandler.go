/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var logger = logging.NewLogger("fabsdk/client")

var lsccFilter = func(ccID string) bool {
	return ccID != "lscc"
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
			requestContext.Error = errors.WithMessage(err, "error getting additional endorsers")
			return
		}

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

func getEndorsers(requestContext *RequestContext, clientContext *ClientContext) (ccCalls []*fab.ChaincodeCall, peers []fab.Peer, err error) {
	var selectionOpts []options.Opt
	if requestContext.SelectionFilter != nil {
		selectionOpts = append(selectionOpts, selectopts.WithPeerFilter(requestContext.SelectionFilter))
	}

	ccCalls = newChaincodeCalls(requestContext.Request)
	peers, err = clientContext.Selection.GetEndorsersForChaincode(ccCalls, selectionOpts...)
	return
}

func getAdditionalEndorsers(requestContext *RequestContext, clientContext *ClientContext, ccCalls []*fab.ChaincodeCall) ([]fab.Peer, error) {
	ccIDs, err := getChaincodes(requestContext.Response.Responses[0])
	if err != nil {
		return nil, err
	}

	additionalCalls := getAdditionalCalls(ccCalls, ccIDs, getCCFilter(requestContext))
	if len(additionalCalls) == 0 {
		return nil, nil
	}

	logger.Debugf("Checking if additional endorsements are required...")
	requestContext.Request.InvocationChain = append(requestContext.Request.InvocationChain, additionalCalls...)

	_, endorsers, err := getEndorsers(requestContext, clientContext)
	if err != nil {
		return nil, err
	}

	var additionalEndorsers []fab.Peer
	for _, endorser := range endorsers {
		if !containsMSP(requestContext.Opts.Targets, endorser.MSPID()) {
			logger.Debugf("Will ask for additional endorsement from [%s] in order to satisfy the chaincode policy", endorser.URL())
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

func getChaincodes(response *fab.TransactionProposalResponse) ([]string, error) {
	rwSets, err := getRWSetsFromProposalResponse(response.ProposalResponse)
	if err != nil {
		return nil, err
	}
	return getNamespaces(rwSets), nil
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

func getNamespaces(rwSets []*rwsetutil.NsRwSet) []string {
	namespaceMap := make(map[string]bool)
	for _, rwSet := range rwSets {
		namespaceMap[rwSet.NameSpace] = true
	}

	var namespaces []string
	for ns := range namespaceMap {
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

func getAdditionalCalls(ccCalls []*fab.ChaincodeCall, namespaces []string, filter CCFilter) []*fab.ChaincodeCall {
	var additionalCalls []*fab.ChaincodeCall
	for _, ccID := range namespaces {
		if !filter(ccID) {
			logger.Debugf("Ignoring chaincode [%s] in the RW set since it was filtered out", ccID)
			continue
		}
		if !containsCC(ccCalls, ccID) {
			logger.Debugf("Found additional chaincode [%s] in the RW set that was not part of the original invocation chain", ccID)
			additionalCalls = append(additionalCalls, &fab.ChaincodeCall{ID: ccID})
		}
	}
	return additionalCalls
}

func containsCC(ccCalls []*fab.ChaincodeCall, ccID string) bool {
	for _, ccCall := range ccCalls {
		if ccCall.ID == ccID {
			return true
		}
	}
	return false
}
