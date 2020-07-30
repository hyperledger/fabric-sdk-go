/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	reqContext "context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	lifecyclepkg "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

//go:generate counterfeiter -o mocklifecycleresource.gen.go -fake-name MockLifecycleResource . lifecycleResource
//go:generate counterfeiter -o mockchannelprovider.gen.go -fake-name MockChannelProvider ../../common/providers/fab ChannelProvider
//go:generate counterfeiter -o mockchannelservice.gen.go -fake-name MockChannelService ../../common/providers/fab ChannelService
//go:generate counterfeiter -o mocktransactor.gen.go -fake-name MockTransactor ../../common/providers/fab Transactor

type lifecycleResource interface {
	Install(reqCtx reqContext.Context, installPkg []byte, targets []fab.ProposalProcessor, opts ...resource.Opt) ([]*resource.LifecycleInstallProposalResponse, error)
	GetInstalledPackage(reqCtx reqContext.Context, packageID string, target fab.ProposalProcessor, opts ...resource.Opt) ([]byte, error)
	QueryInstalled(reqCtx reqContext.Context, target fab.ProposalProcessor, opts ...resource.Opt) (*resource.LifecycleQueryInstalledCCResponse, error)
	QueryApproved(reqCtx reqContext.Context, channelID string, req *resource.QueryApprovedChaincodeRequest, target fab.ProposalProcessor, opts ...resource.Opt) (*resource.LifecycleQueryApprovedCCResponse, error)
	CreateApproveProposal(txh fab.TransactionHeader, req *resource.ApproveChaincodeRequest) (*fab.TransactionProposal, error)
	CreateCheckCommitReadinessProposal(txh fab.TransactionHeader, req *resource.CheckChaincodeCommitReadinessRequest) (*fab.TransactionProposal, error)
	CreateCommitProposal(txh fab.TransactionHeader, req *resource.CommitChaincodeRequest) (*fab.TransactionProposal, error)
	CreateQueryCommittedProposal(txh fab.TransactionHeader, req *resource.QueryCommittedChaincodesRequest) (*fab.TransactionProposal, error)
	UnmarshalApplicationPolicy(policyBytes []byte) (*common.SignaturePolicyEnvelope, string, error)
}

type targetProvider func(channelID string, opts requestOptions) ([]fab.Peer, error)
type signatureVerifier func(channelService fab.ChannelService, txProposalResponse []*fab.TransactionProposalResponse) error
type committer func(eventService fab.EventService, tp *fab.TransactionProposal, txProposalResponse []*fab.TransactionProposalResponse, transac fab.Transactor, reqCtx reqContext.Context) (fab.TransactionID, error)
type protoUnmarshaller func(buf []byte, pb proto.Message) error

type lifecycleProcessor struct {
	lifecycleResource
	ctx                  context.Client
	getCCProposalTargets targetProvider
	verifyTPSignature    signatureVerifier
	commitTransaction    committer
	protoUnmarshal       protoUnmarshaller
}

func newLifecycleProcessor(ctx context.Client, targetProvider targetProvider, signatureVerifier signatureVerifier, committer committer) *lifecycleProcessor {
	return &lifecycleProcessor{
		lifecycleResource:    resource.NewLifecycle(),
		ctx:                  ctx,
		getCCProposalTargets: targetProvider,
		verifyTPSignature:    signatureVerifier,
		commitTransaction:    committer,
		protoUnmarshal:       proto.Unmarshal,
	}
}

func (p *lifecycleProcessor) install(reqCtx reqContext.Context, installPkg []byte, targets []fab.Peer) ([]LifecycleInstallCCResponse, error) {
	tpResponses, err := p.Install(reqCtx, installPkg, peer.PeersToTxnProcessors(targets))
	if err != nil {
		return nil, err
	}

	var responses []LifecycleInstallCCResponse
	for _, v := range tpResponses {
		logger.Debugf("Install chaincode endorser '%s' returned response status: %d", v.Endorser, v.Status)

		response := LifecycleInstallCCResponse{
			Target:    v.Endorser,
			Status:    v.Status,
			PackageID: v.PackageId,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (p *lifecycleProcessor) queryInstalled(reqCtx reqContext.Context, target fab.Peer) ([]LifecycleInstalledCC, error) {
	r, err := p.QueryInstalled(reqCtx, target)
	if err != nil {
		return nil, errors.WithMessage(err, "querying for installed chaincodes failed")
	}

	logger.Debugf("Query installed chaincodes endorser '%s' returned ProposalResponse status:%v", r.Endorser, r.Status)

	return p.toInstalledChaincodes(r.InstalledChaincodes), nil
}

func (p *lifecycleProcessor) approve(reqCtx reqContext.Context, channelID string, req LifecycleApproveCCRequest, opts requestOptions) (fab.TransactionID, error) {
	if err := p.verifyApproveParams(channelID, req); err != nil {
		return fab.EmptyTransactionID, err
	}

	targets, channelService, transactor, txh, err := p.prepare(reqCtx, channelID, opts)
	if err != nil {
		return fab.EmptyTransactionID, err
	}

	eventService, err := channelService.EventService()
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "unable to get event service")
	}

	var acr = resource.ApproveChaincodeRequest(req)

	tp, err := p.lifecycleResource.CreateApproveProposal(txh, &acr)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "creation of approve chaincode proposal failed")
	}

	// Process and send transaction proposal
	txProposalResponse, err := transactor.SendTransactionProposal(tp, peersToTxnProcessors(targets))
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "sending approve transaction proposal failed")
	}

	// Verify signature(s)
	err = p.verifyTPSignature(channelService, txProposalResponse)
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "sending approve transaction proposal failed to verify signature")
	}

	// send transaction and check event
	return p.commitTransaction(eventService, tp, txProposalResponse, transactor, reqCtx)
}

func (p *lifecycleProcessor) queryApproved(reqCtx reqContext.Context, channelID string, req LifecycleQueryApprovedCCRequest, target fab.Peer) (LifecycleApprovedChaincodeDefinition, error) {
	if err := p.verifyQueryApprovedParams(channelID, req); err != nil {
		return LifecycleApprovedChaincodeDefinition{}, err
	}

	r := &resource.QueryApprovedChaincodeRequest{
		Name:     req.Name,
		Sequence: req.Sequence,
	}

	tpr, err := p.QueryApproved(reqCtx, channelID, r, target)
	if err != nil {
		return LifecycleApprovedChaincodeDefinition{}, errors.WithMessage(err, "querying for installed chaincode failed")
	}

	logger.Debugf("Query approved chaincodes endorser '%s' returned ProposalResponse status:%v", tpr.Endorser, tpr.Status)

	return LifecycleApprovedChaincodeDefinition(*tpr.ApprovedChaincode), nil
}

func (p *lifecycleProcessor) checkCommitReadiness(reqCtx reqContext.Context, channelID string, req LifecycleCheckCCCommitReadinessRequest, opts requestOptions) (LifecycleCheckCCCommitReadinessResponse, error) {
	if err := p.verifyCheckCommitReadinessParams(channelID, req); err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, err
	}

	targets, channelService, transactor, txh, err := p.prepare(reqCtx, channelID, opts)
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, err
	}

	var ccr = resource.CheckChaincodeCommitReadinessRequest(req)

	tp, err := p.CreateCheckCommitReadinessProposal(txh, &ccr)
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, errors.WithMessage(err, "creation of check chaincode commit readiness proposal failed")
	}

	txProposalResponse, err := transactor.SendTransactionProposal(tp, peersToTxnProcessors(targets))
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, errors.WithMessage(err, "sending approve transaction proposal failed")
	}

	if len(txProposalResponse) == 0 {
		return LifecycleCheckCCCommitReadinessResponse{}, errors.New("no responses")
	}

	err = p.verifyTPSignature(channelService, txProposalResponse)
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, errors.WithMessage(err, "sending approve transaction proposal failed to verify signature")
	}

	err = p.verifyResponsesMatch(txProposalResponse,
		func(payload []byte) (proto.Message, error) {
			result := &lb.CheckCommitReadinessResult{}
			return result, proto.Unmarshal(payload, result)
		},
	)
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, err
	}

	result := &lb.CheckCommitReadinessResult{}
	err = p.protoUnmarshal(txProposalResponse[0].Response.Payload, result)
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	return LifecycleCheckCCCommitReadinessResponse{
		Approvals: result.Approvals,
	}, nil
}

func (p *lifecycleProcessor) commit(reqCtx reqContext.Context, channelID string, req LifecycleCommitCCRequest, opts requestOptions) (fab.TransactionID, error) {
	if err := p.verifyCommitParams(channelID, req); err != nil {
		return fab.EmptyTransactionID, err
	}

	targets, channelService, transactor, txh, err := p.prepare(reqCtx, channelID, opts)
	if err != nil {
		return fab.EmptyTransactionID, err
	}

	eventService, err := channelService.EventService()
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "unable to get event service")
	}

	var cr = resource.CommitChaincodeRequest(req)

	tp, err := p.CreateCommitProposal(txh, &cr)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "creation of commit chaincode proposal failed")
	}

	// Process and send transaction proposal
	txProposalResponse, err := transactor.SendTransactionProposal(tp, peersToTxnProcessors(targets))
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "sending commit transaction proposal failed")
	}

	// Verify signature(s)
	err = p.verifyTPSignature(channelService, txProposalResponse)
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "sending commit transaction proposal failed to verify signature")
	}

	// send transaction and check event
	return p.commitTransaction(eventService, tp, txProposalResponse, transactor, reqCtx)
}

func (p *lifecycleProcessor) queryCommitted(reqCtx reqContext.Context, channelID string, req LifecycleQueryCommittedCCRequest, opts requestOptions) ([]LifecycleChaincodeDefinition, error) {
	if channelID == "" {
		return nil, errors.New("channel ID is required")
	}

	targets, channelService, transactor, txh, err := p.prepare(reqCtx, channelID, opts)
	if err != nil {
		return nil, err
	}

	tp, err := p.CreateQueryCommittedProposal(txh, &resource.QueryCommittedChaincodesRequest{Name: req.Name})
	if err != nil {
		return nil, errors.WithMessage(err, "creation of query chaincode definitions proposal failed")
	}

	txProposalResponse, err := transactor.SendTransactionProposal(tp, peersToTxnProcessors(targets))
	if err != nil {
		return nil, errors.WithMessage(err, "sending commit transaction proposal failed")
	}

	if len(txProposalResponse) == 0 {
		return nil, errors.New("no responses")
	}

	err = p.verifyTPSignature(channelService, txProposalResponse)
	if err != nil {
		return nil, errors.WithMessage(err, "sending query committed transaction proposal failed to verify signature")
	}

	err = p.verifyResponsesMatch(txProposalResponse,
		func(payload []byte) (proto.Message, error) {
			return unmarshalCCDefResults(req.Name, payload)
		},
	)
	if err != nil {
		return nil, err
	}

	payload := txProposalResponse[0].Response.Payload

	if req.Name != "" {
		def, err := p.unmarshalChaincodeDefinition(req.Name, payload)
		if err != nil {
			return nil, err
		}

		return []LifecycleChaincodeDefinition{def}, nil
	}

	return p.unmarshalChaincodeDefinitions(payload)
}

func (p *lifecycleProcessor) prepare(reqCtx reqContext.Context, channelID string, opts requestOptions) ([]fab.Peer, fab.ChannelService, fab.Transactor, *txn.TransactionHeader, error) {
	targets, err := p.getCCProposalTargets(channelID, opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	channelService, err := p.ctx.ChannelProvider().ChannelService(p.ctx, channelID)
	if err != nil {
		return nil, nil, nil, nil, errors.WithMessage(err, "Unable to get channel service")
	}

	transactor, err := channelService.Transactor(reqCtx)
	if err != nil {
		return nil, nil, nil, nil, errors.WithMessage(err, "get channel transactor failed")
	}

	txh, err := txn.NewHeader(p.ctx, channelID)
	if err != nil {
		return nil, nil, nil, nil, errors.WithMessage(err, "create transaction ID failed")
	}

	return targets, channelService, transactor, txh, nil
}

func (p *lifecycleProcessor) adjustTargetsForInstall(targets []fab.Peer, req LifecycleInstallCCRequest, retry retry.Opts, parentReqCtx reqContext.Context) ([]fab.Peer, multi.Errors) {
	errs := multi.Errors{}

	// Targets will be adjusted if cc has already been installed
	newTargets := make([]fab.Peer, 0)
	for _, target := range targets {
		reqCtx, cancel := contextImpl.NewRequest(p.ctx, contextImpl.WithTimeoutType(fab.PeerResponse), contextImpl.WithParent(parentReqCtx))
		defer cancel()

		installed, err := p.isInstalled(reqCtx, req, target, retry)
		if err != nil {
			// Add to errors with unable to verify error message
			errs = append(errs, errors.Errorf("unable to verify if cc is installed on %s. Got error: %s", target.URL(), err))
			continue
		}

		if !installed {
			// Not installed - add for processing
			newTargets = append(newTargets, target)
		}
	}

	return newTargets, errs
}

func (p *lifecycleProcessor) verifyInstallParams(req LifecycleInstallCCRequest) error {
	if req.Label == "" {
		return errors.New("label is required")
	}

	if len(req.Package) == 0 {
		return errors.New("package is required")
	}

	return nil
}

func (p *lifecycleProcessor) verifyApproveParams(channelID string, req LifecycleApproveCCRequest) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.Version == "" {
		return errors.New("version is required")
	}

	return nil
}

func (p *lifecycleProcessor) verifyQueryApprovedParams(channelID string, req LifecycleQueryApprovedCCRequest) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

func (p *lifecycleProcessor) verifyCheckCommitReadinessParams(channelID string, req LifecycleCheckCCCommitReadinessRequest) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.Version == "" {
		return errors.New("version is required")
	}

	return nil
}

func (p *lifecycleProcessor) verifyCommitParams(channelID string, req LifecycleCommitCCRequest) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.Version == "" {
		return errors.New("version is required")
	}

	return nil
}

func (p *lifecycleProcessor) isInstalled(reqCtx reqContext.Context, req LifecycleInstallCCRequest, peer fab.ProposalProcessor, retryOpts retry.Opts) (bool, error) {
	packageID := lifecyclepkg.ComputePackageID(req.Label, req.Package)

	_, err := p.GetInstalledPackage(reqCtx, packageID, peer, resource.WithRetry(retryOpts))
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("chaincode install package '%s' not found", packageID)) {
			logger.Debugf("Chaincode package [%s] is not installed", packageID)

			return false, nil
		}

		return false, err
	}

	logger.Debugf("Chaincode package [%s] has already been installed", packageID)

	return true, nil
}

func (p *lifecycleProcessor) toInstalledChaincodes(installedChaincodes []resource.LifecycleInstalledCC) []LifecycleInstalledCC {
	ccs := make([]LifecycleInstalledCC, len(installedChaincodes))
	for i, ic := range installedChaincodes {
		refsByChannelID := make(map[string][]CCReference)
		for channelID, chaincodes := range ic.References {
			refs := make([]CCReference, len(chaincodes))
			for j, cc := range chaincodes {
				refs[j] = CCReference{
					Name:    cc.Name,
					Version: cc.Version,
				}
			}

			refsByChannelID[channelID] = refs
		}
		ccs[i] = LifecycleInstalledCC{
			PackageID:  ic.PackageID,
			Label:      ic.Label,
			References: refsByChannelID,
		}
	}

	return ccs
}

func (p *lifecycleProcessor) unmarshalChaincodeDefinition(name string, payload []byte) (LifecycleChaincodeDefinition, error) {
	result := &lb.QueryChaincodeDefinitionResult{}
	err := p.protoUnmarshal(payload, result)
	if err != nil {
		return LifecycleChaincodeDefinition{}, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	var collConfig []*pb.CollectionConfig
	if result.Collections != nil {
		collConfig = result.Collections.Config
	}

	signaturePolicy, channelConfigPolicy, err := p.lifecycleResource.UnmarshalApplicationPolicy(result.ValidationParameter)
	if err != nil {
		return LifecycleChaincodeDefinition{}, err
	}

	return LifecycleChaincodeDefinition{
		Name:                name,
		Version:             result.Version,
		Sequence:            result.Sequence,
		EndorsementPlugin:   result.EndorsementPlugin,
		ValidationPlugin:    result.ValidationPlugin,
		SignaturePolicy:     signaturePolicy,
		ChannelConfigPolicy: channelConfigPolicy,
		CollectionConfig:    collConfig,
		InitRequired:        result.InitRequired,
		Approvals:           result.Approvals,
	}, nil
}

func (p *lifecycleProcessor) unmarshalChaincodeDefinitions(payload []byte) ([]LifecycleChaincodeDefinition, error) {
	result := &lb.QueryChaincodeDefinitionsResult{}
	err := p.protoUnmarshal(payload, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	results := make([]LifecycleChaincodeDefinition, len(result.ChaincodeDefinitions))

	for i, def := range result.ChaincodeDefinitions {
		var collConfig []*pb.CollectionConfig
		if def.Collections != nil {
			collConfig = def.Collections.Config
		}

		signaturePolicy, channelConfigPolicy, err := p.lifecycleResource.UnmarshalApplicationPolicy(def.ValidationParameter)
		if err != nil {
			return nil, err
		}

		results[i] = LifecycleChaincodeDefinition{
			Name:                def.Name,
			Version:             def.Version,
			Sequence:            def.Sequence,
			EndorsementPlugin:   def.EndorsementPlugin,
			ValidationPlugin:    def.ValidationPlugin,
			SignaturePolicy:     signaturePolicy,
			ChannelConfigPolicy: channelConfigPolicy,
			CollectionConfig:    collConfig,
			InitRequired:        def.InitRequired,
		}
	}

	return results, nil
}

type unmarshaller func(payload []byte) (proto.Message, error)

// verifyResponsesMatch ensures that the payload in all of the responses are the same
func (p *lifecycleProcessor) verifyResponsesMatch(responses []*fab.TransactionProposalResponse, unmarshal unmarshaller) error {
	var lastStatus int32
	var lastResponse proto.Message

	for _, r := range responses {
		m, err := unmarshal(r.Response.Payload)
		if err != nil {
			return err
		}

		if lastResponse != nil {
			if lastStatus != r.Response.Status {
				return errors.Errorf("status in responses from endorsers do not match: [%d] and [%d]", lastStatus, r.Response.Status)
			}

			if !proto.Equal(lastResponse, m) {
				return errors.Errorf("responses from endorsers do not match: [%+v] and [%+v]", lastResponse, m)
			}
		}

		lastResponse = m
		lastStatus = r.Response.Status
	}

	return nil
}

func unmarshalCCDefResults(name string, payload []byte) (proto.Message, error) {
	if name != "" {
		result := &lb.QueryChaincodeDefinitionResult{}
		return result, proto.Unmarshal(payload, result)
	}

	result := &lb.QueryChaincodeDefinitionsResult{}
	return result, proto.Unmarshal(payload, result)
}
