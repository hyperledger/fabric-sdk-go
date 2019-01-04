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
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/common/privdata"
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

// DirEnumerator enumerates directories
type DirEnumerator func(string) ([]os.FileInfo, error)

// ChaincodeExtractor extracts chaincode from a given path
type ChaincodeExtractor func(ccname string, ccversion string, path string) (CCPackage, error)

// ccInfoFSStorageMgr is the storage manager used either by the cache or if the
// cache is bypassed
var ccInfoFSProvider = &CCInfoFSImpl{}

// ccInfoCache is the cache instance itself

// CCContext pass this around instead of string of args
type CCContext struct {
	// Name chaincode name
	Name string

	// Version used to construct the chaincode image and register
	Version string
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

// ChaincodeContainerInfo is yet another synonym for the data required to start/stop a chaincode.
type ChaincodeContainerInfo struct {
	Name        string
	Version     string
	Path        string
	Type        string
	CodePackage []byte

	// ContainerType is not a great name, but 'DOCKER' and 'SYSTEM' are the valid types
	ContainerType string
}

// TransactionParams are parameters which are tied to a particular transaction
// and which are required for invoking chaincode.
type TransactionParams struct {
	TxID                 string
	ChannelID            string
	SignedProp           *pb.SignedProposal
	Proposal             *pb.Proposal
	TXSimulator          ledger.TxSimulator
	HistoryQueryExecutor ledger.HistoryQueryExecutor
	CollectionStore      privdata.CollectionStore
	IsInitTransaction    bool

	// this is additional data passed to the chaincode
	ProposalDecorations map[string][]byte
}

// ChaincodeProvider provides an abstraction layer that is
// used for different packages to interact with code in the
// chaincode package without importing it; more methods
// should be added below if necessary
type ChaincodeProvider interface {
	// Execute executes a standard chaincode invocation for a chaincode and an input
	Execute(txParams *TransactionParams, cccid *CCContext, input *pb.ChaincodeInput) (*pb.Response, *pb.ChaincodeEvent, error)
	// ExecuteLegacyInit is a special case for executing chaincode deployment specs,
	// which are not already in the LSCC, needed for old lifecycle
	ExecuteLegacyInit(txParams *TransactionParams, cccid *CCContext, spec *pb.ChaincodeDeploymentSpec) (*pb.Response, *pb.ChaincodeEvent, error)
	// Stop stops the chaincode give
	Stop(ccci *ChaincodeContainerInfo) error
}
