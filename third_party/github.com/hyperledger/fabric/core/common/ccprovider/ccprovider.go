/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package ccprovider

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var ccproviderLogger = flogging.MustGetLogger("ccprovider")

var chaincodeInstallPath string

// CCPackage encapsulates a chaincode package which can be
//    raw ChaincodeDeploymentSpec
//    SignedChaincodeDeploymentSpec
// Attempt to keep the interface at a level with minimal
// interface for possible generalization.
type CCPackage interface {
	//InitFromBuffer initialize the package from bytes
	InitFromBuffer(buf []byte) (*ChaincodeData, error)

	// InitFromFS gets the chaincode from the filesystem (includes the raw bytes too)
	InitFromFS(ccname string, ccversion string) ([]byte, *pb.ChaincodeDeploymentSpec, error)

	// PutChaincodeToFS writes the chaincode to the filesystem
	PutChaincodeToFS() error

	// GetDepSpec gets the ChaincodeDeploymentSpec from the package
	GetDepSpec() *pb.ChaincodeDeploymentSpec

	// GetDepSpecBytes gets the serialized ChaincodeDeploymentSpec from the package
	GetDepSpecBytes() []byte

	// ValidateCC validates and returns the chaincode deployment spec corresponding to
	// ChaincodeData. The validation is based on the metadata from ChaincodeData
	// One use of this method is to validate the chaincode before launching
	ValidateCC(ccdata *ChaincodeData) error

	// GetPackageObject gets the object as a proto.Message
	GetPackageObject() proto.Message

	// GetChaincodeData gets the ChaincodeData
	GetChaincodeData() *ChaincodeData

	// GetId gets the fingerprint of the chaincode based on package computation
	GetId() []byte
}

type CCCacheSupport interface {
	// GetChaincode is needed by the cache to get chaincode data
	GetChaincode(ccname string, ccversion string) (CCPackage, error)
}

// CCInfoFSImpl provides the implementation for CC on the FS and the access to it
// It implements CCCacheSupport
type CCInfoFSImpl struct{}

// The following lines create the cache of CCPackage data that sits
// on top of the file system and avoids a trip to the file system
// every time. The cache is disabled by default and only enabled
// if EnableCCInfoCache is called. This is an unfortunate hack
// required by some legacy tests that remove chaincode packages
// from the file system as a means of simulating particular test
// conditions. This way of testing is incompatible with the
// immutable nature of chaincode packages that is assumed by hlf v1
// and implemented by this cache. For this reason, tests are for now
// allowed to run with the cache disabled (unless they enable it)
// until a later time in which they are fixed. The peer process on
// the other hand requires the benefits of this cache and therefore
// enables it.
// TODO: (post v1) enable cache by default as soon as https://jira.hyperledger.org/browse/FAB-3785 is completed

// ccInfoFSStorageMgr is the storage manager used either by the cache or if the
// cache is bypassed
var ccInfoFSProvider = &CCInfoFSImpl{}

// ccInfoCache is the cache instance itself

// ccInfoCacheEnabled keeps track of whether the cache is enable
// (it is disabled by default)
var ccInfoCacheEnabled bool

// CCContext pass this around instead of string of args
type CCContext struct {
	// ChainID chain id
	ChainID string

	// Name chaincode name
	Name string

	// Version used to construct the chaincode image and register
	Version string

	// TxID is the transaction id for the proposal (if any)
	TxID string

	// Syscc is this a system chaincode
	Syscc bool

	// SignedProposal for this invoke (if any) this is kept here for access
	// control and in case we need to pass something from this to the chaincode
	SignedProposal *pb.SignedProposal

	// Proposal for this invoke (if any) this is kept here just in case we need to
	// pass something from this to the chaincode
	Proposal *pb.Proposal

	// canonicalName is not set but computed
	canonicalName string

	// this is additional data passed to the chaincode
	ProposalDecorations map[string][]byte
}

func (cccid *CCContext) String() string {
	return fmt.Sprintf("chain=%s,chaincode=%s,version=%s,txid=%s,syscc=%t,proposal=%p,canname=%s",
		cccid.ChainID, cccid.Name, cccid.Version, cccid.TxID, cccid.Syscc, cccid.Proposal, cccid.canonicalName)
}

//-------- ChaincodeDefinition - interface for ChaincodeData ------
// ChaincodeDefinition describes all of the necessary information for a peer to decide whether to endorse
// a proposal and whether to validate a transaction, for a particular chaincode.
type ChaincodeDefinition interface {
	// CCName returns the name of this chaincode (the name it was put in the ChaincodeRegistry with).
	CCName() string

	// Hash returns the hash of the chaincode.
	Hash() []byte

	// CCVersion returns the version of the chaincode.
	CCVersion() string

	// Validation returns how to validate transactions for this chaincode.
	// The string returned is the name of the validation method (usually 'vscc')
	// and the bytes returned are the argument to the validation (in the case of
	// 'vscc', this is a marshaled pb.VSCCArgs message).
	Validation() (string, []byte)

	// Endorsement returns how to endorse proposals for this chaincode.
	// The string returns is the name of the endorsement method (usually 'escc').
	Endorsement() string
}

//-------- ChaincodeData is stored on the LSCC -------

// ChaincodeData defines the datastructure for chaincodes to be serialized by proto
// Type provides an additional check by directing to use a specific package after instantiation
// Data is Type specifc (see CDSPackage and SignedCDSPackage)
type ChaincodeData struct {
	// Name of the chaincode
	Name string `protobuf:"bytes,1,opt,name=name"`

	// Version of the chaincode
	Version string `protobuf:"bytes,2,opt,name=version"`

	// Escc for the chaincode instance
	Escc string `protobuf:"bytes,3,opt,name=escc"`

	// Vscc for the chaincode instance
	Vscc string `protobuf:"bytes,4,opt,name=vscc"`

	// Policy endorsement policy for the chaincode instance
	Policy []byte `protobuf:"bytes,5,opt,name=policy,proto3"`

	// Data data specific to the package
	Data []byte `protobuf:"bytes,6,opt,name=data,proto3"`

	// Id of the chaincode that's the unique fingerprint for the CC This is not
	// currently used anywhere but serves as a good eyecatcher
	Id []byte `protobuf:"bytes,7,opt,name=id,proto3"`

	// InstantiationPolicy for the chaincode
	InstantiationPolicy []byte `protobuf:"bytes,8,opt,name=instantiation_policy,proto3"`
}

// implement functions needed from proto.Message for proto's mar/unmarshal functions

// Reset resets
func (cd *ChaincodeData) Reset() { *cd = ChaincodeData{} }

// String converts to string
func (cd *ChaincodeData) String() string { return proto.CompactTextString(cd) }

// ProtoMessage just exists to make proto happy
func (*ChaincodeData) ProtoMessage() {}

// ChaincodeSpecGetter normalizes getting a chaincode spec from an
// ChaincodeInvocationSpec or a ChaincodeDeploymentSpec.
type ChaincodeSpecGetter interface {
	GetChaincodeSpec() *pb.ChaincodeSpec
}

// ChaincodeProvider provides an abstraction layer that is
// used for different packages to interact with code in the
// chaincode package without importing it; more methods
// should be added below if necessary
type ChaincodeProvider interface {
	// GetContext returns a ledger context and a tx simulator; it's the
	// caller's responsability to release the simulator by calling its
	// done method once it is no longer useful
	GetContext(ledger ledger.PeerLedger, txid string) (context.Context, ledger.TxSimulator, error)
	// ExecuteChaincode executes the chaincode given context and args
	ExecuteChaincode(ctxt context.Context, cccid *CCContext, args [][]byte) (*pb.Response, *pb.ChaincodeEvent, error)
	// Execute executes the chaincode given context and spec (invocation or deploy)
	Execute(ctxt context.Context, cccid *CCContext, spec ChaincodeSpecGetter) (*pb.Response, *pb.ChaincodeEvent, error)
	// Stop stops the chaincode given context and deployment spec
	Stop(ctxt context.Context, cccid *CCContext, spec *pb.ChaincodeDeploymentSpec) error
}
