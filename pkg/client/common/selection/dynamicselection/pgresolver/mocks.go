/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import (
	"github.com/golang/protobuf/proto"
	common "github.com/hyperledger/fabric-protos-go/common"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/pkg/errors"
)

// NewPrincipal creates a new MSPPrincipal
func NewPrincipal(name string, classification mb.MSPPrincipal_Classification) (*mb.MSPPrincipal, error) {
	member1Role, err := proto.Marshal(&mb.MSPRole{Role: mb.MSPRole_MEMBER, MspIdentifier: name})
	if err != nil {
		return nil, errors.WithMessage(err, "Error marshal MSPRole")
	}
	return &mb.MSPPrincipal{
		PrincipalClassification: classification,
		Principal:               member1Role}, nil
}

// NewSignedByPolicy creates a SignaturePolicy at the given index
func NewSignedByPolicy(index int32) *common.SignaturePolicy {
	return &common.SignaturePolicy{
		Type: &common.SignaturePolicy_SignedBy{
			SignedBy: index,
		}}
}

// NewNOutOfPolicy creates an NOutOf signature policy
func NewNOutOfPolicy(n int32, signedBy ...*common.SignaturePolicy) *common.SignaturePolicy {
	return &common.SignaturePolicy{
		Type: &common.SignaturePolicy_NOutOf_{
			NOutOf: &common.SignaturePolicy_NOutOf{
				N:     n,
				Rules: signedBy,
			}}}
}

// GetPolicies creates a set of 'signed by' signature policies and corresponding identities for the given set of MSP IDs
func GetPolicies(mspIDs ...string) (signedBy []*common.SignaturePolicy, identities []*mb.MSPPrincipal, err error) {
	for i, mspID := range mspIDs {
		signedBy = append(signedBy, NewSignedByPolicy(int32(i)))
		principal, err := NewPrincipal(mspID, mb.MSPPrincipal_ROLE)
		if err != nil {
			return nil, nil, err
		}
		identities = append(identities, principal)
	}
	return
}
