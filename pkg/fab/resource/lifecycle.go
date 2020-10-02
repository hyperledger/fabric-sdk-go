/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	reqContext "context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

const (
	lifecycleCC = "_lifecycle"

	lifecycleInstallFuncName                  = "InstallChaincode"
	lifecycleQueryInstalledChaincodesFunc     = "QueryInstalledChaincodes"
	lifecycleGetInstalledChaincodePackageFunc = "GetInstalledChaincodePackage"
	lifecycleApproveChaincodeFuncName         = "ApproveChaincodeDefinitionForMyOrg"
	lifecycleQueryApprovedCCDefinitionFunc    = "QueryApprovedChaincodeDefinition"
	lifecycleCheckCommitReadinessFuncName     = "CheckCommitReadiness"
	lifecycleCommitFuncName                   = "CommitChaincodeDefinition"
	lifecycleQueryChaincodeDefinitionFunc     = "QueryChaincodeDefinition"
	lifecycleQueryChaincodeDefinitionsFunc    = "QueryChaincodeDefinitions"
)

// ApproveChaincodeRequest contains the parameters required to approve a chaincode
type ApproveChaincodeRequest struct {
	Name                string
	Version             string
	PackageID           string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     *common.SignaturePolicyEnvelope
	ChannelConfigPolicy string
	CollectionConfig    []*pb.CollectionConfig
	InitRequired        bool
}

// QueryApprovedChaincodeRequest contains the parameters for an approved chaincode query
type QueryApprovedChaincodeRequest struct {
	Name     string
	Sequence int64
}

// CommitChaincodeRequest contains the parameters for a commit chaincode request
type CommitChaincodeRequest struct {
	Name                string
	Version             string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     *common.SignaturePolicyEnvelope
	ChannelConfigPolicy string
	CollectionConfig    []*pb.CollectionConfig
	InitRequired        bool
}

// CheckChaincodeCommitReadinessRequest contains the parameters for checking the 'commit readiness' of a chaincode
type CheckChaincodeCommitReadinessRequest struct {
	Name                string
	Version             string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     *common.SignaturePolicyEnvelope
	ChannelConfigPolicy string
	CollectionConfig    []*pb.CollectionConfig
	InitRequired        bool
}

// QueryCommittedChaincodesRequest contains the parameters to query committed chaincodes.
// If name is not provided then all committed chaincodes on the given channel are returned,
// otherwise only the chaincode with the given name is returned.
type QueryCommittedChaincodesRequest struct {
	Name string
}

type protoMarshaller func(pb proto.Message) ([]byte, error)
type protoUnmarshaller func(buf []byte, pb proto.Message) error
type contextProvider func(ctx reqContext.Context) (context.Client, bool)
type txnHeaderProvider func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error)

// Lifecycle implements chaincode lifecycle operations
type Lifecycle struct {
	protoMarshal   protoMarshaller
	protoUnmarshal protoUnmarshaller
	newContext     contextProvider
	newTxnHeader   txnHeaderProvider
}

// NewLifecycle returns a Lifecycle resource implementation that handles all chaincode lifecycle functions
func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		protoMarshal:   proto.Marshal,
		protoUnmarshal: proto.Unmarshal,
		newContext:     contextImpl.RequestClientContext,
		newTxnHeader:   txn.NewHeader,
	}
}

// Install installs a chaincode package
func (lc *Lifecycle) Install(reqCtx reqContext.Context, installPkg []byte, targets []fab.ProposalProcessor, opts ...Opt) ([]*LifecycleInstallProposalResponse, error) {
	if len(installPkg) == 0 {
		return nil, errors.New("chaincode package is required")
	}

	ctx, ok := lc.newContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn header")
	}

	txh, err := lc.newTxnHeader(ctx, fab.SystemChannel)
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	prop, err := lc.createInstallProposal(txh, installPkg)
	if err != nil {
		return nil, errors.WithMessage(err, "creation of install chaincode proposal failed")
	}

	optionsValue := getOpts(opts...)

	resp, err := retry.NewInvoker(retry.New(optionsValue.retry)).Invoke(
		func() (interface{}, error) {
			return txn.SendProposal(reqCtx, prop, targets)
		},
	)
	if err != nil {
		return nil, err
	}

	response := resp.([]*fab.TransactionProposalResponse)
	installResponse := make([]*LifecycleInstallProposalResponse, len(response))
	for i, r := range response {
		ir := &lb.InstallChaincodeResult{}
		err = lc.protoUnmarshal(r.Response.Payload, ir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal install chaincode result")
		}

		installResponse[i] = &LifecycleInstallProposalResponse{
			TransactionProposalResponse: r,
			InstallChaincodeResult:      ir,
		}
	}

	return installResponse, nil
}

// QueryInstalled returns information about the installed chaincodes on a given peer.
func (lc *Lifecycle) QueryInstalled(reqCtx reqContext.Context, target fab.ProposalProcessor, opts ...Opt) (*LifecycleQueryInstalledCCResponse, error) {
	ctx, ok := lc.newContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn header")
	}

	txh, err := lc.newTxnHeader(ctx, fab.SystemChannel)
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	prop, err := lc.createQueryInstalledProposal(txh)
	if err != nil {
		return nil, errors.WithMessage(err, "creation of query installed chaincodes proposal failed")
	}

	optionsValue := getOpts(opts...)

	resp, err := retry.NewInvoker(retry.New(optionsValue.retry)).Invoke(
		func() (interface{}, error) {
			return txn.SendProposal(reqCtx, prop, []fab.ProposalProcessor{target})
		},
	)
	if err != nil {
		return nil, err
	}

	tpResponses := resp.([]*fab.TransactionProposalResponse)

	r := tpResponses[0]
	logger.Debugf("Query installed chaincodes endorser '%s' returned ProposalResponse status:%v", r.Endorser, r.Status)

	qicr := &lb.QueryInstalledChaincodesResult{}
	err = lc.protoUnmarshal(r.Response.Payload, qicr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	return &LifecycleQueryInstalledCCResponse{
		TransactionProposalResponse: r,
		InstalledChaincodes:         toInstalledChaincodes(qicr.InstalledChaincodes),
	}, nil
}

// GetInstalledPackage returns the installed chaincode package for the given package ID
func (lc *Lifecycle) GetInstalledPackage(reqCtx reqContext.Context, packageID string, target fab.ProposalProcessor, opts ...Opt) ([]byte, error) {
	ctx, ok := lc.newContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn header")
	}

	txh, err := lc.newTxnHeader(ctx, fab.SystemChannel)
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	prop, err := lc.createGetInstalledPackageProposal(txh, packageID)
	if err != nil {
		return nil, errors.WithMessage(err, "creation of get installed chaincode package proposal failed")
	}

	optionsValue := getOpts(opts...)

	resp, err := retry.NewInvoker(retry.New(optionsValue.retry)).Invoke(
		func() (interface{}, error) {
			return txn.SendProposal(reqCtx, prop, []fab.ProposalProcessor{target})
		},
	)
	if err != nil {
		return nil, err
	}

	tpResponses := resp.([]*fab.TransactionProposalResponse)

	r := tpResponses[0]

	logger.Debugf("Get installed chaincode package endorser '%s' returned ProposalResponse status:%v", r.Endorser, r.Status)

	qicr := &lb.GetInstalledChaincodePackageResult{}
	err = lc.protoUnmarshal(r.Response.Payload, qicr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	return qicr.ChaincodeInstallPackage, nil
}

// QueryApproved returns information about the approved chaincode
func (lc *Lifecycle) QueryApproved(reqCtx reqContext.Context, channelID string, req *QueryApprovedChaincodeRequest, target fab.ProposalProcessor, opts ...Opt) (*LifecycleQueryApprovedCCResponse, error) {
	ctx, ok := lc.newContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn header")
	}

	txh, err := lc.newTxnHeader(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	prop, err := lc.createQueryApprovedDefinitionProposal(txh, req)
	if err != nil {
		return nil, errors.WithMessage(err, "creation of query approved chaincodes proposal failed")
	}

	optionsValue := getOpts(opts...)

	resp, err := retry.NewInvoker(retry.New(optionsValue.retry)).Invoke(
		func() (interface{}, error) {
			return txn.SendProposal(reqCtx, prop, []fab.ProposalProcessor{target})
		},
	)
	if err != nil {
		return nil, err
	}

	tpResponses := resp.([]*fab.TransactionProposalResponse)

	tpr := tpResponses[0]

	logger.Debugf("Query approved chaincodes endorser '%s' returned ProposalResponse status:%v", tpr.Endorser, tpr.Status)

	result := &lb.QueryApprovedChaincodeDefinitionResult{}
	err = lc.protoUnmarshal(tpr.Response.Payload, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	approvedCC, err := lc.toApprovedChaincodes(req.Name, result)
	if err != nil {
		return nil, err
	}

	return &LifecycleQueryApprovedCCResponse{
		TransactionProposalResponse: tpr,
		ApprovedChaincode:           approvedCC,
	}, nil
}

func (lc *Lifecycle) createInstallProposal(txh fab.TransactionHeader, installPkg []byte) (*fab.TransactionProposal, error) {
	cir, err := lc.createInstallRequest(installPkg)
	if err != nil {
		return nil, errors.WithMessage(err, "creating lscc install invocation request failed")
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

func (lc *Lifecycle) createInstallRequest(installPkg []byte) (fab.ChaincodeInvokeRequest, error) {
	installChaincodeArgs := &lb.InstallChaincodeArgs{
		ChaincodeInstallPackage: installPkg,
	}

	installChaincodeArgsBytes, err := lc.protoMarshal(installChaincodeArgs)
	if err != nil {
		return fab.ChaincodeInvokeRequest{}, errors.Wrap(err, "failed to marshal InstallChaincodeArgs")
	}

	return fab.ChaincodeInvokeRequest{
		ChaincodeID: lifecycleCC,
		Fcn:         lifecycleInstallFuncName,
		Args:        [][]byte{installChaincodeArgsBytes},
	}, nil
}

func (lc *Lifecycle) createQueryInstalledProposal(txh fab.TransactionHeader) (*fab.TransactionProposal, error) {
	args := &lb.QueryInstalledChaincodesArgs{}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal InstallChaincodeArgs")
	}

	return txn.CreateChaincodeInvokeProposal(txh,
		fab.ChaincodeInvokeRequest{
			ChaincodeID: lifecycleCC,
			Fcn:         lifecycleQueryInstalledChaincodesFunc,
			Args:        [][]byte{argsBytes},
		},
	)
}

func (lc *Lifecycle) createGetInstalledPackageProposal(txh fab.TransactionHeader, packageID string) (*fab.TransactionProposal, error) {
	args := &lb.GetInstalledChaincodePackageArgs{
		PackageId: packageID,
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal GetInstalledChaincodePackageArgs")
	}

	return txn.CreateChaincodeInvokeProposal(txh,
		fab.ChaincodeInvokeRequest{
			ChaincodeID: lifecycleCC,
			Fcn:         lifecycleGetInstalledChaincodePackageFunc,
			Args:        [][]byte{argsBytes},
		},
	)
}

// CreateApproveProposal creates a proposal to query approved chaincodes
func (lc *Lifecycle) CreateApproveProposal(txh fab.TransactionHeader, req *ApproveChaincodeRequest) (*fab.TransactionProposal, error) {
	var ccsrc *lb.ChaincodeSource
	if req.PackageID != "" {
		ccsrc = &lb.ChaincodeSource{
			Type: &lb.ChaincodeSource_LocalPackage{
				LocalPackage: &lb.ChaincodeSource_Local{
					PackageId: req.PackageID,
				},
			},
		}
	} else {
		ccsrc = &lb.ChaincodeSource{
			Type: &lb.ChaincodeSource_Unavailable_{
				Unavailable: &lb.ChaincodeSource_Unavailable{},
			},
		}
	}

	policyBytes, err := lc.MarshalApplicationPolicy(req.SignaturePolicy, req.ChannelConfigPolicy)
	if err != nil {
		return nil, errors.WithMessage(err, "create application policy failed")
	}

	args := &lb.ApproveChaincodeDefinitionForMyOrgArgs{
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		ValidationParameter: policyBytes,
		InitRequired:        req.InitRequired,
		Collections:         &pb.CollectionConfigPackage{Config: req.CollectionConfig},
		Source:              ccsrc,
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, err
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lifecycleCC,
		Fcn:         lifecycleApproveChaincodeFuncName,
		Args:        [][]byte{argsBytes},
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

// CreateCommitProposal creates a proposal to commit a chaincode
func (lc *Lifecycle) CreateCommitProposal(txh fab.TransactionHeader, req *CommitChaincodeRequest) (*fab.TransactionProposal, error) {
	policyBytes, err := lc.MarshalApplicationPolicy(req.SignaturePolicy, req.ChannelConfigPolicy)
	if err != nil {
		return nil, errors.WithMessage(err, "create application policy failed")
	}

	args := &lb.CommitChaincodeDefinitionArgs{
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		ValidationParameter: policyBytes,
		InitRequired:        req.InitRequired,
		Collections:         &pb.CollectionConfigPackage{Config: req.CollectionConfig},
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, err
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lifecycleCC,
		Fcn:         lifecycleCommitFuncName,
		Args:        [][]byte{argsBytes},
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

// CreateCheckCommitReadinessProposal creates a propoposal to check 'commit readiness' of a chaincode
func (lc *Lifecycle) CreateCheckCommitReadinessProposal(txh fab.TransactionHeader, req *CheckChaincodeCommitReadinessRequest) (*fab.TransactionProposal, error) {
	policyBytes, err := lc.MarshalApplicationPolicy(req.SignaturePolicy, req.ChannelConfigPolicy)
	if err != nil {
		return nil, errors.WithMessage(err, "create application policy failed")
	}

	args := &lb.CheckCommitReadinessArgs{
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		ValidationParameter: policyBytes,
		InitRequired:        req.InitRequired,
		Collections:         &pb.CollectionConfigPackage{Config: req.CollectionConfig},
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, err
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lifecycleCC,
		Fcn:         lifecycleCheckCommitReadinessFuncName,
		Args:        [][]byte{argsBytes},
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

// CreateQueryCommittedProposal creates a propoposal to query for committed chaincodes. If the chaincode name is provided
// in the request then the proposal will contain a query for a single chaincode, otherwise all committed chaincodes on the
// chainnel will be queried.
func (lc *Lifecycle) CreateQueryCommittedProposal(txh fab.TransactionHeader, req *QueryCommittedChaincodesRequest) (*fab.TransactionProposal, error) {
	var function string
	var args proto.Message

	if req.Name != "" {
		function = lifecycleQueryChaincodeDefinitionFunc
		args = &lb.QueryChaincodeDefinitionArgs{
			Name: req.Name,
		}
	} else {
		function = lifecycleQueryChaincodeDefinitionsFunc
		args = &lb.QueryChaincodeDefinitionsArgs{}
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, err
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lifecycleCC,
		Fcn:         function,
		Args:        [][]byte{argsBytes},
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

func (lc *Lifecycle) createQueryApprovedDefinitionProposal(txh fab.TransactionHeader, req *QueryApprovedChaincodeRequest) (*fab.TransactionProposal, error) {
	args := &lb.QueryApprovedChaincodeDefinitionArgs{
		Name:     req.Name,
		Sequence: req.Sequence,
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, err
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lifecycleCC,
		Fcn:         lifecycleQueryApprovedCCDefinitionFunc,
		Args:        [][]byte{argsBytes},
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

// MarshalApplicationPolicy marshals the given signature or channel config policy into an ApplicationPolicy payload
func (lc *Lifecycle) MarshalApplicationPolicy(signaturePolicy *common.SignaturePolicyEnvelope, channelConfigPolicy string) ([]byte, error) {
	if signaturePolicy == nil && channelConfigPolicy == "" {
		return nil, nil
	}

	if signaturePolicy != nil && channelConfigPolicy != "" {
		return nil, errors.New("cannot specify both signature policy and channel config policy")
	}

	var applicationPolicy *pb.ApplicationPolicy
	if signaturePolicy != nil {
		applicationPolicy = &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_SignaturePolicy{
				SignaturePolicy: signaturePolicy,
			},
		}
	}

	if channelConfigPolicy != "" {
		applicationPolicy = &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_ChannelConfigPolicyReference{
				ChannelConfigPolicyReference: channelConfigPolicy,
			},
		}
	}

	policyBytes, err := lc.protoMarshal(applicationPolicy)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to marshal application policy")
	}

	return policyBytes, nil
}

// UnmarshalApplicationPolicy unmarshals the policy baytes and returns either a signature policy or a channel config policy.
func (lc *Lifecycle) UnmarshalApplicationPolicy(policyBytes []byte) (*common.SignaturePolicyEnvelope, string, error) {
	applicationPolicy := &pb.ApplicationPolicy{}
	err := lc.protoUnmarshal(policyBytes, applicationPolicy)
	if err != nil {
		return nil, "", errors.WithMessage(err, "failed to unmarshal application policy")
	}

	switch policy := applicationPolicy.Type.(type) {
	case *pb.ApplicationPolicy_SignaturePolicy:
		return policy.SignaturePolicy, "", nil
	case *pb.ApplicationPolicy_ChannelConfigPolicyReference:
		return nil, policy.ChannelConfigPolicyReference, nil
	default:
		return nil, "", errors.Errorf("unsupported policy type %T", policy)
	}
}

func (lc *Lifecycle) toApprovedChaincodes(ccName string, result *lb.QueryApprovedChaincodeDefinitionResult) (*LifecycleApprovedCC, error) {
	var collConfig []*pb.CollectionConfig
	if result.Collections != nil {
		collConfig = result.Collections.Config
	}

	var packageID string
	if result.Source != nil {
		switch source := result.Source.Type.(type) {
		case *lb.ChaincodeSource_LocalPackage:
			packageID = source.LocalPackage.PackageId
		case *lb.ChaincodeSource_Unavailable_:
		}
	}

	signaturePolicy, channelConfigPolicy, err := lc.UnmarshalApplicationPolicy(result.ValidationParameter)
	if err != nil {
		return nil, err
	}

	return &LifecycleApprovedCC{
		Name:                ccName,
		Version:             result.Version,
		Sequence:            result.Sequence,
		EndorsementPlugin:   result.EndorsementPlugin,
		ValidationPlugin:    result.ValidationPlugin,
		SignaturePolicy:     signaturePolicy,
		ChannelConfigPolicy: channelConfigPolicy,
		CollectionConfig:    collConfig,
		InitRequired:        result.InitRequired,
		PackageID:           packageID,
	}, nil
}

func toInstalledChaincodes(installedChaincodes []*lb.QueryInstalledChaincodesResult_InstalledChaincode) []LifecycleInstalledCC {
	result := make([]LifecycleInstalledCC, len(installedChaincodes))
	for i, ic := range installedChaincodes {
		refsByChannelID := make(map[string][]CCReference)
		for channelID, chaincodes := range ic.References {
			refs := make([]CCReference, len(chaincodes.Chaincodes))
			for j, cc := range chaincodes.Chaincodes {
				refs[j] = CCReference{
					Name:    cc.Name,
					Version: cc.Version,
				}
			}

			refsByChannelID[channelID] = refs
		}
		result[i] = LifecycleInstalledCC{
			PackageID:  ic.PackageId,
			Label:      ic.Label,
			References: refsByChannelID,
		}
	}

	return result
}
