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
)

//go:generate counterfeiter -o mocklifecycleresource.gen.go -fake-name MockLifecycleResource . lifecycleResource

type lifecycleResource interface {
	Install(reqCtx reqContext.Context, installPkg []byte, targets []fab.ProposalProcessor, opts ...resource.Opt) ([]*resource.LifecycleInstallProposalResponse, error)
	GetInstalledPackage(reqCtx reqContext.Context, packageID string, target fab.ProposalProcessor, opts ...resource.Opt) ([]byte, error)
}

type lifecycleProcessor struct {
	lifecycleResource
	ctx context.Client
}

func newLifecycleProcessor(ctx context.Client) *lifecycleProcessor {
	return &lifecycleProcessor{
		lifecycleResource: resource.NewLifecycle(),
		ctx:               ctx,
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
