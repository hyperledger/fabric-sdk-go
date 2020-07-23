/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	reqContext "context"
	"fmt"
	"strings"

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

type lifecycleResource interface {
	Install(reqCtx reqContext.Context, installPkg []byte, targets []fab.ProposalProcessor, opts ...resource.Opt) ([]*resource.LifecycleInstallProposalResponse, error)
	GetInstalledPackage(reqCtx reqContext.Context, packageID string, target fab.ProposalProcessor, opts ...resource.Opt) ([]byte, error)
	QueryInstalled(reqCtx reqContext.Context, target fab.ProposalProcessor, opts ...resource.Opt) (*resource.LifecycleQueryInstalledCCResponse, error)
	QueryApproved(reqCtx reqContext.Context, channelID string, req *resource.QueryApprovedChaincodeRequest, target fab.ProposalProcessor, opts ...resource.Opt) (*resource.LifecycleQueryApprovedCCResponse, error)
	CreateApproveProposal(txh fab.TransactionHeader, req *resource.ApproveChaincodeRequest) (*fab.TransactionProposal, error)
}

type targetProvider func(channelID string, opts requestOptions) ([]fab.Peer, error)
type signatureVerifier func(channelService fab.ChannelService, txProposalResponse []*fab.TransactionProposalResponse) error
type committer func(eventService fab.EventService, tp *fab.TransactionProposal, txProposalResponse []*fab.TransactionProposalResponse, transac fab.Transactor, reqCtx reqContext.Context) (fab.TransactionID, error)

type lifecycleProcessor struct {
	lifecycleResource
	ctx                  context.Client
	getCCProposalTargets targetProvider
	verifyTPSignature    signatureVerifier
	commitTransaction    committer
}

func newLifecycleProcessor(ctx context.Client, targetProvider targetProvider, signatureVerifier signatureVerifier, committer committer) *lifecycleProcessor {
	return &lifecycleProcessor{
		lifecycleResource:    resource.NewLifecycle(),
		ctx:                  ctx,
		getCCProposalTargets: targetProvider,
		verifyTPSignature:    signatureVerifier,
		commitTransaction:    committer,
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

	targets, err := p.getCCProposalTargets(channelID, opts)
	if err != nil {
		return fab.EmptyTransactionID, err
	}

	// Get transactor on the channel to create and send the deploy proposal
	channelService, err := p.ctx.ChannelProvider().ChannelService(p.ctx, channelID)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "Unable to get channel service")
	}

	transactor, err := channelService.Transactor(reqCtx)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "get channel transactor failed")
	}

	eventService, err := channelService.EventService()
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "unable to get event service")
	}

	txh, err := txn.NewHeader(p.ctx, channelID)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "create transaction ID failed")
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
