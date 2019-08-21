/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	common "github.com/hyperledger/fabric-protos-go/common"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

const loggerModule = "fabsdk/client"

var logger = logging.NewLogger(loggerModule)

// GroupRetriever is a function that returns groups of peers
type GroupRetriever func(peerRetriever MSPPeerRetriever) (GroupOfGroups, error)

type peerGroupResolver struct {
	groupRetriever GroupRetriever
	lbp            LoadBalancePolicy
}

// NewRoundRobinPeerGroupResolver returns a PeerGroupResolver that chooses peers in a round-robin fashion
func NewRoundRobinPeerGroupResolver(sigPolicyEnv *common.SignaturePolicyEnvelope) (PeerGroupResolver, error) {
	groupRetriever, err := CompileSignaturePolicy(sigPolicyEnv)
	if err != nil {
		return nil, errors.WithMessage(err, "error evaluating signature policy")
	}
	return NewPeerGroupResolver(groupRetriever, NewRoundRobinLBP())
}

// NewRandomPeerGroupResolver returns a PeerGroupResolver that chooses peers in a round-robin fashion
func NewRandomPeerGroupResolver(sigPolicyEnv *common.SignaturePolicyEnvelope) (PeerGroupResolver, error) {
	groupRetriever, err := CompileSignaturePolicy(sigPolicyEnv)
	if err != nil {
		return nil, errors.WithMessage(err, "error evaluating signature policy")
	}
	return NewPeerGroupResolver(groupRetriever, NewRandomLBP())
}

// NewPeerGroupResolver returns a new PeerGroupResolver
func NewPeerGroupResolver(groupRetriever GroupRetriever, lbp LoadBalancePolicy) (PeerGroupResolver, error) {
	return &peerGroupResolver{
		groupRetriever: groupRetriever,
		lbp:            lbp,
	}, nil
}

func (c *peerGroupResolver) Resolve(peers []fab.Peer) (PeerGroup, error) {
	peerRetriever := func(mspID string) []fab.Peer {
		var mspPeers []fab.Peer
		for _, peer := range peers {
			if mspID == "" || peer.MSPID() == mspID {
				mspPeers = append(mspPeers, peer)
			}
		}
		return mspPeers
	}

	peerGroups, err := c.getPeerGroups(peerRetriever)
	if err != nil {
		return nil, err
	}

	if logging.IsEnabledFor(loggerModule, logging.DEBUG) {
		var s string
		if len(peerGroups) == 0 {
			s = "  ***** No Available Peer Groups "
		} else {
			s = "  ***** Available Peer Groups: "
			for i, grp := range peerGroups {
				s += fmt.Sprintf("%d - %+v", i, grp)
				if i+1 < len(peerGroups) {
					s += fmt.Sprintf(" OR ")
				}
			}
			s += fmt.Sprintf(" ")
		}
		logger.Debugf(s)
	}

	return c.lbp.Choose(peerGroups), nil
}

func (c *peerGroupResolver) getPeerGroups(peerRetriever MSPPeerRetriever) ([]PeerGroup, error) {
	groupHierarchy, err := c.groupRetriever(peerRetriever)
	if err != nil {
		return nil, err
	}

	logger.Debugf("***** Policy: %s", groupHierarchy)

	mspGroups := groupHierarchy.Reduce()

	if logging.IsEnabledFor(loggerModule, logging.DEBUG) {
		s := " ***** Org Groups: "
		for i, g := range mspGroups {
			s += fmt.Sprintf("%+v", g)
			if i+1 < len(mspGroups) {
				s += fmt.Sprintf("  OR ")
			}
		}
		s += fmt.Sprintf(" ")
		logger.Debugf(s)
	}

	var allPeerGroups []PeerGroup
	for _, g := range mspGroups {
		allPeerGroups = append(allPeerGroups, mustGetPeerGroups(g)...)
	}
	return allPeerGroups, nil
}

func mustGetPeerGroups(group Group) []PeerGroup {
	items := group.Items()
	if len(items) == 0 {
		return nil
	}

	if _, ok := items[0].(Group); !ok {
		group = NewGroup([]Item{group})
	}

	groups := make([]Group, len(group.Items()))
	for i, item := range group.Items() {
		if grp, ok := item.(PeerGroup); ok {
			groups[i] = grp
		} else {
			panic(fmt.Sprintf("unexpected: expecting item to be a PeerGroup but found: %s", reflect.TypeOf(item)))
		}
	}

	andedGroups := and(groups)
	peerGroups := make([]PeerGroup, len(andedGroups))
	for i, g := range andedGroups {
		peerGroups[i] = mustGetPeerGroup(g)
	}

	return peerGroups
}

func mustGetPeerGroup(g Group) PeerGroup {
	if pg, ok := g.(PeerGroup); ok {
		return pg
	}

	var peers []fab.Peer
	for _, item := range g.Items() {
		if pg, ok := item.(fab.Peer); ok {
			peers = append(peers, pg)
		} else {
			panic(fmt.Sprintf("expecting item to be a Peer but found: %s", reflect.TypeOf(item)))
		}
	}
	return NewPeerGroup(peers...)
}

// CompileSignaturePolicy compiles the given signature policy and returns a GroupRetriever
func CompileSignaturePolicy(sigPolicyEnv *common.SignaturePolicyEnvelope) (GroupRetriever, error) {
	compiler := &signaturePolicyCompiler{}
	return compiler.Compile(sigPolicyEnv)
}

type signaturePolicyCompiler struct {
}

func (c *signaturePolicyCompiler) Compile(sigPolicyEnv *common.SignaturePolicyEnvelope) (GroupRetriever, error) {
	policFunc, err := c.compile(sigPolicyEnv.Rule, sigPolicyEnv.Identities)
	if err != nil {
		return nil, errors.WithMessage(err, "error compiling chaincode signature policy")
	}
	return policFunc, nil
}

func (c *signaturePolicyCompiler) compile(sigPolicy *common.SignaturePolicy, identities []*mb.MSPPrincipal) (GroupRetriever, error) {
	if sigPolicy == nil {
		return nil, errors.New("nil signature policy")
	}

	switch t := sigPolicy.Type.(type) {

	case *common.SignaturePolicy_SignedBy:

		return signaturePolicySignedBy(t, identities)

	case *common.SignaturePolicy_NOutOf_:

		return c.signaturePolicyNOutOf(t, identities)

	default:
		errMsg := fmt.Sprintf("unsupported signature policy type: %v", t)
		return nil, errors.New(errMsg)

	}
}

func (c *signaturePolicyCompiler) signaturePolicyNOutOf(t *common.SignaturePolicy_NOutOf_, identities []*mb.MSPPrincipal) (GroupRetriever, error) {

	nOutOfPolicy := t.NOutOf
	if nOutOfPolicy.N == 0 {
		return signaturePolicySignedByAny()
	}

	var pfuncs []GroupRetriever
	for _, policy := range nOutOfPolicy.Rules {
		f, err := c.compile(policy, identities)
		if err != nil {
			return nil, err
		}
		pfuncs = append(pfuncs, f)
	}

	return func(peerRetriever MSPPeerRetriever) (GroupOfGroups, error) {
		var groups []Group
		for _, f := range pfuncs {
			grps, err := f(peerRetriever)
			if err != nil {
				return nil, err
			}
			groups = append(groups, grps)
		}

		itemGroups, err := NewGroupOfGroups(groups).Nof(nOutOfPolicy.N)
		if err != nil {
			return nil, err
		}

		return itemGroups, nil
	}, nil
}

func signaturePolicySignedBy(t *common.SignaturePolicy_SignedBy, identities []*mb.MSPPrincipal) (GroupRetriever, error) {
	return func(peerRetriever MSPPeerRetriever) (GroupOfGroups, error) {
		mspID, err := mspPrincipalToString(identities[t.SignedBy])
		if err != nil {
			return nil, errors.WithMessage(err, "error getting MSP ID from MSP principal")
		}
		return NewGroupOfGroups([]Group{NewMSPPeerGroup(mspID, peerRetriever)}), nil
	}, nil
}

func signaturePolicySignedByAny() (GroupRetriever, error) {
	return func(peerRetriever MSPPeerRetriever) (GroupOfGroups, error) {
		return NewGroupOfGroups([]Group{NewMSPPeerGroup("", peerRetriever)}), nil
	}, nil
}

func mspPrincipalToString(principal *mb.MSPPrincipal) (string, error) {
	switch principal.PrincipalClassification {
	case mb.MSPPrincipal_ROLE:
		// Principal contains the msp role
		mspRole := &mb.MSPRole{}
		err := proto.Unmarshal(principal.Principal, mspRole)
		if err != nil {
			return "", errors.WithMessage(err, "unmarshal of principal failed")
		}
		return mspRole.MspIdentifier, nil

	case mb.MSPPrincipal_ORGANIZATION_UNIT:
		// Principal contains the OrganizationUnit
		unit := &mb.OrganizationUnit{}
		err := proto.Unmarshal(principal.Principal, unit)
		if err != nil {
			return "", errors.WithMessage(err, "unmarshal of principal failed")
		}
		return unit.MspIdentifier, nil

	case mb.MSPPrincipal_IDENTITY:
		// TODO: Do we need to support this?
		errMsg := fmt.Sprintf("unsupported PrincipalClassification type: %s", reflect.TypeOf(principal.PrincipalClassification))
		return "", errors.New(errMsg)

	default:
		errMsg := fmt.Sprintf("unknown PrincipalClassification type: %s", reflect.TypeOf(principal.PrincipalClassification))
		return "", errors.New(errMsg)
	}
}
