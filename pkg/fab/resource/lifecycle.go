/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	reqContext "context"

	"github.com/golang/protobuf/proto"
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
	lifecycleGetInstalledChaincodePackageFunc = "GetInstalledChaincodePackage"
	lifecycleQueryCommittedFuncName           = "QueryChaincodeDefinition"
)

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

// QueryCommitted returns the chaincode definition for the given ccName
func (lc *Lifecycle) QueryCommitted(reqCtx reqContext.Context, ccName string, channelName string, target fab.ProposalProcessor, opts ...Opt) ([]*LifecycleQueryCommittedResponse, error) {
	ctx, ok := lc.newContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn header")
	}

	txh, err := lc.newTxnHeader(ctx, channelName)
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	prop, err := lc.createQueryCommittedProposal(txh, ccName)
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

	logger.Debugf("QueryCommitted endorser '%s' returned ProposalResponse status:%v", r.Endorser, r.Status)

	queryResponse := make([]*LifecycleQueryCommittedResponse, 0)
	if ccName != "" {
		result := &lb.QueryChaincodeDefinitionResult{}
		err := proto.Unmarshal(r.Response.Payload, result)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
		}
		queryResponse = append(queryResponse, &LifecycleQueryCommittedResponse{
			Name:                ccName,
			Sequence:            result.Sequence,
			Version:             result.Version,
			EndorsementPlugin:   result.EndorsementPlugin,
			ValidationPlugin:    result.ValidationPlugin,
			ValidationParameter: result.ValidationParameter,
			Collections:         result.Collections,
			InitRequired:        result.InitRequired,
		})
		return queryResponse, nil
	}

	result := &lb.QueryChaincodeDefinitionsResult{}
	err = lc.protoUnmarshal(r.Response.Payload, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal proposal response's response payload")
	}

	for _, cd := range result.ChaincodeDefinitions {
		queryResponse = append(queryResponse, &LifecycleQueryCommittedResponse{
			Name:                cd.Name,
			Sequence:            cd.Sequence,
			Version:             cd.Version,
			EndorsementPlugin:   cd.EndorsementPlugin,
			ValidationPlugin:    cd.ValidationPlugin,
			ValidationParameter: cd.ValidationParameter,
			Collections:         cd.Collections,
			InitRequired:        cd.InitRequired,
		})
	}
	return queryResponse, nil
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

func (lc *Lifecycle) createQueryCommittedProposal(txh fab.TransactionHeader, ccName string) (*fab.TransactionProposal, error) {
	var function string
	var args proto.Message

	if ccName != "" {
		function = lifecycleQueryCommittedFuncName
		args = &lb.QueryChaincodeDefinitionArgs{
			Name: ccName,
		}
	} else {
		function = lifecycleQueryCommittedFuncName
		args = &lb.QueryChaincodeDefinitionsArgs{}
	}

	argsBytes, err := lc.protoMarshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal QueryCommittedArgs")
	}

	return txn.CreateChaincodeInvokeProposal(txh,
		fab.ChaincodeInvokeRequest{
			ChaincodeID: lifecycleCC,
			Fcn:         function,
			Args:        [][]byte{argsBytes},
		},
	)
}
