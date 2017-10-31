/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	sdkApi "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	common "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
)

var logger = logging.NewLogger("fabric_sdk_go")

type peerGroupResolver struct {
	mspGroups []Group
	lbp       LoadBalancePolicy
}

// NewRoundRobinPeerGroupResolver returns a PeerGroupResolver that chooses peers in a round-robin fashion
func NewRoundRobinPeerGroupResolver(sigPolicyEnv *common.SignaturePolicyEnvelope, peerRetriever PeerRetriever) (PeerGroupResolver, error) {
	compiler := NewSignaturePolicyCompiler(peerRetriever)
	groupHierarchy, err := compiler.Compile(sigPolicyEnv)
	if err != nil {
		return nil, errors.WithMessage(err, "error evaluating signature policy")
	}
	return NewPeerGroupResolver(groupHierarchy, NewRoundRobinLBP())
}

// NewRandomPeerGroupResolver returns a PeerGroupResolver that chooses peers in a round-robin fashion
func NewRandomPeerGroupResolver(sigPolicyEnv *common.SignaturePolicyEnvelope, peerRetriever PeerRetriever) (PeerGroupResolver, error) {
	compiler := NewSignaturePolicyCompiler(peerRetriever)
	groupHierarchy, err := compiler.Compile(sigPolicyEnv)
	if err != nil {
		return nil, errors.WithMessage(err, "error evaluating signature policy")
	}
	return NewPeerGroupResolver(groupHierarchy, NewRandomLBP())
}

// NewPeerGroupResolver returns a new PeerGroupResolver
func NewPeerGroupResolver(groupHierarchy GroupOfGroups, lbp LoadBalancePolicy) (PeerGroupResolver, error) {

	logger.Debugf("\n***** Policy: %s\n", groupHierarchy)

	mspGroups := groupHierarchy.Reduce()

	s := "\n***** Org Groups:\n"
	for i, g := range mspGroups {
		s += fmt.Sprintf("%s", g)
		if i+1 < len(mspGroups) {
			s += fmt.Sprintf("  OR\n")
		}
	}
	s += fmt.Sprintf("\n")
	logger.Debugf(s)

	return &peerGroupResolver{
		mspGroups: mspGroups,
		lbp:       lbp,
	}, nil
}

func (c *peerGroupResolver) Resolve() PeerGroup {
	peerGroups := c.getPeerGroups()

	s := ""
	if len(peerGroups) == 0 {
		s = "\n\n***** No Available Peer Groups\n"
	} else {
		s = "\n\n***** Available Peer Groups:\n"
		for i, grp := range peerGroups {
			s += fmt.Sprintf("%d - %s", i, grp)
			if i+1 < len(peerGroups) {
				s += fmt.Sprintf(" OR\n")
			}
		}
		s += fmt.Sprintf("\n")
	}

	logger.Debugf(s)

	return c.lbp.Choose(peerGroups)
}

func (c *peerGroupResolver) getPeerGroups() []PeerGroup {
	var allPeerGroups []PeerGroup
	for _, g := range c.mspGroups {
		for _, pg := range mustGetPeerGroups(g) {
			allPeerGroups = append(allPeerGroups, pg)
		}
	}
	return allPeerGroups
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

	var peers []sdkApi.Peer
	for _, item := range g.Items() {
		if pg, ok := item.(sdkApi.Peer); ok {
			peers = append(peers, pg)
		} else {
			panic(fmt.Sprintf("expecting item to be a Peer but found: %s", reflect.TypeOf(item)))
		}
	}
	return NewPeerGroup(peers...)
}

// NewSignaturePolicyCompiler returns a new PolicyCompiler
func NewSignaturePolicyCompiler(peerRetriever PeerRetriever) SignaturePolicyCompiler {
	return &signaturePolicyCompiler{
		peerRetriever: peerRetriever,
	}
}

type signaturePolicyCompiler struct {
	peerRetriever PeerRetriever
}

func (c *signaturePolicyCompiler) Compile(sigPolicyEnv *common.SignaturePolicyEnvelope) (GroupOfGroups, error) {
	policFunc, err := c.compile(sigPolicyEnv.Rule, sigPolicyEnv.Identities)
	if err != nil {
		return nil, errors.WithMessage(err, "error compiling chaincode signature policy")
	}
	return policFunc()
}

func (c *signaturePolicyCompiler) compile(sigPolicy *common.SignaturePolicy, identities []*mb.MSPPrincipal) (SignaturePolicyFunc, error) {
	if sigPolicy == nil {
		return nil, errors.New("nil signature policy")
	}

	switch t := sigPolicy.Type.(type) {
	case *common.SignaturePolicy_SignedBy:
		return func() (GroupOfGroups, error) {
			mspID, err := mspPrincipalToString(identities[t.SignedBy])
			if err != nil {
				return nil, errors.WithMessage(err, "error getting MSP ID from MSP principal")
			}
			return NewGroupOfGroups([]Group{NewMSPPeerGroup(mspID, c.peerRetriever)}), nil
		}, nil

	case *common.SignaturePolicy_NOutOf_:
		nOutOfPolicy := t.NOutOf
		var pfuncs []SignaturePolicyFunc
		for _, policy := range nOutOfPolicy.Rules {
			f, err := c.compile(policy, identities)
			if err != nil {
				return nil, err
			}
			pfuncs = append(pfuncs, f)
		}
		return func() (GroupOfGroups, error) {
			var groups []Group
			for _, f := range pfuncs {
				grps, err := f()
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

	default:
		errMsg := fmt.Sprintf("unsupported signature policy type: %v", t)
		return nil, errors.New(errMsg)
	}
}

func mspPrincipalToString(principal *mb.MSPPrincipal) (string, error) {
	switch principal.PrincipalClassification {
	case mb.MSPPrincipal_ROLE:
		// Principal contains the msp role
		mspRole := &mb.MSPRole{}
		proto.Unmarshal(principal.Principal, mspRole)
		return mspRole.MspIdentifier, nil

	case mb.MSPPrincipal_ORGANIZATION_UNIT:
		// Principal contains the OrganizationUnit
		unit := &mb.OrganizationUnit{}
		proto.Unmarshal(principal.Principal, unit)
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
