/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package policies

import (
	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
)

const (
	// Path separator is used to separate policy names in paths
	PathSeparator = "/"

	// ChannelPrefix is used in the path of standard channel policy managers
	ChannelPrefix = "Channel"

	// ApplicationPrefix is used in the path of standard application policy paths
	ApplicationPrefix = "Application"

	// OrdererPrefix is used in the path of standard orderer policy paths
	OrdererPrefix = "Orderer"

	// ChannelReaders is the label for the channel's readers policy (encompassing both orderer and application readers)
	ChannelReaders = PathSeparator + ChannelPrefix + PathSeparator + "Readers"

	// ChannelWriters is the label for the channel's writers policy (encompassing both orderer and application writers)
	ChannelWriters = PathSeparator + ChannelPrefix + PathSeparator + "Writers"

	// ChannelApplicationReaders is the label for the channel's application readers policy
	ChannelApplicationReaders = PathSeparator + ChannelPrefix + PathSeparator + ApplicationPrefix + PathSeparator + "Readers"

	// ChannelApplicationWriters is the label for the channel's application writers policy
	ChannelApplicationWriters = PathSeparator + ChannelPrefix + PathSeparator + ApplicationPrefix + PathSeparator + "Writers"

	// ChannelApplicationAdmins is the label for the channel's application admin policy
	ChannelApplicationAdmins = PathSeparator + ChannelPrefix + PathSeparator + ApplicationPrefix + PathSeparator + "Admins"

	// BlockValidation is the label for the policy which should validate the block signatures for the channel
	BlockValidation = PathSeparator + ChannelPrefix + PathSeparator + OrdererPrefix + PathSeparator + "BlockValidation"

	// ChannelOrdererAdmins is the label for the channel's orderer admin policy
	ChannelOrdererAdmins = PathSeparator + ChannelPrefix + PathSeparator + OrdererPrefix + PathSeparator + "Admins"

	// ChannelOrdererWriters is the label for the channel's orderer writers policy
	ChannelOrdererWriters = PathSeparator + ChannelPrefix + PathSeparator + OrdererPrefix + PathSeparator + "Writers"

	// ChannelOrdererReaders is the label for the channel's orderer readers policy
	ChannelOrdererReaders = PathSeparator + ChannelPrefix + PathSeparator + OrdererPrefix + PathSeparator + "Readers"
)

var logger = flogging.MustGetLogger("policies")

// PrincipalSet is a collection of MSPPrincipals
type PrincipalSet []*msp.MSPPrincipal

// PrincipalSets aggregates PrincipalSets
type PrincipalSets []PrincipalSet

// Converter represents a policy
// which may be translated into a SignaturePolicyEnvelope
type Converter interface {
	Convert() (*cb.SignaturePolicyEnvelope, error)
}

// Policy is used to determine if a signature is valid
type Policy interface {
	// EvaluateSignedData takes a set of SignedData and evaluates whether
	// 1) the signatures are valid over the related message
	// 2) the signing identities satisfy the policy
	EvaluateSignedData(signatureSet []*protoutil.SignedData) error
}

// InquireablePolicy is a Policy that one can inquire
type InquireablePolicy interface {
	// SatisfiedBy returns a slice of PrincipalSets that each of them
	// satisfies the policy.
	SatisfiedBy() []PrincipalSet
}

// Manager is a read only subset of the policy ManagerImpl
type Manager interface {
	// GetPolicy returns a policy and true if it was the policy requested, or false if it is the default policy
	GetPolicy(id string) (Policy, bool)

	// Manager returns the sub-policy manager for a given path and whether it exists
	Manager(path []string) (Manager, bool)
}

// Provider provides the backing implementation of a policy
type Provider interface {
	// NewPolicy creates a new policy based on the policy bytes
	NewPolicy(data []byte) (Policy, proto.Message, error)
}

// ChannelPolicyManagerGetter is a support interface
// to get access to the policy manager of a given channel
type ChannelPolicyManagerGetter interface {
	// Returns the policy manager associated with the specified channel.
	Manager(channelID string) Manager
}

// PolicyManagerGetterFunc is a function adapater for ChannelPolicyManagerGetter.
type PolicyManagerGetterFunc func(channelID string) Manager

// ManagerImpl is an implementation of Manager and configtx.ConfigHandler
// In general, it should only be referenced as an Impl for the configtx.ConfigManager
type ManagerImpl struct {
	path     string // The group level path
	Policies map[string]Policy
	managers map[string]*ManagerImpl
}

type rejectPolicy string

type PolicyLogger struct {
	Policy     Policy
	policyName string
}
