/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/pkg/errors"
)

// GenerateExamplePvtID supplies a chaincode name for example_pvt_cc
func GenerateExamplePvtID(randomize bool) string {
	const (
		chaincodeName = "example_pvt_cc"
	)

	suffix := "0"
	if randomize {
		suffix = GenerateRandomID()
	}

	return fmt.Sprintf("%s_%s%s", chaincodeName, metadata.TestRunID, suffix)
}

// GenerateExampleID supplies a chaincode name for example_cc
func GenerateExampleID(randomize bool) string {
	const (
		chaincodeName = "example_cc"
	)

	suffix := "0"
	if randomize {
		suffix = GenerateRandomID()
	}

	return fmt.Sprintf("%s_0%s%s", chaincodeName, metadata.TestRunID, suffix)
}

// PrepareExampleCC install and instantiate using resource management client
func PrepareExampleCC(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, chaincodeID string) error {
	const (
		ccPath    = "github.com/example_cc"
		ccVersion = "v0"
		channelID = "mychannel"
	)

	instantiated, err := queryInstantiatedCC(sdk, user, orgName, channelID, chaincodeID, ccVersion, false)
	if err != nil {
		return errors.WithMessage(err, "Querying for instantiated status failed")
	}

	if !instantiated {
		fmt.Printf("Installing and instantiating example chaincode...")
		start := time.Now()

		_, err := InstallAndInstantiateCC(sdk, user, orgName, chaincodeID, ccPath, ccVersion, GetDeployPath(), initArgs)
		if err != nil {
			return errors.WithMessage(err, "Installing or instantiating example chaincode failed")
		}

		t := time.Now()
		elapsed := t.Sub(start)
		fmt.Printf("Done [%d ms]\n", elapsed/time.Millisecond)
	} else {
		err := resetExampleCC(sdk, user, orgName, channelID, chaincodeID, resetArgs)
		if err != nil {
			return errors.WithMessage(err, "Resetting example chaincode failed")
		}
	}

	// TODO query all peers to wait for availability

	return nil
}

func resetExampleCC(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, channelID string, chainCodeID string, args [][]byte) error {
	clientContext := sdk.ChannelContext(channelID, user, fabsdk.WithOrg(orgName))

	client, err := channel.New(clientContext)
	if err != nil {
		return errors.WithMessage(err, "Creating channel client failed")
	}

	req := channel.Request{
		ChaincodeID: chainCodeID,
		Fcn:         "reset",
		Args:        args,
	}

	_, err = client.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		return errors.WithMessage(err, "Reset invocation failed")
	}

	return nil
}

func queryInstantiatedCC(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, channelID, ccName, ccVersion string, transientRetry bool) (bool, error) {
	clientContext := sdk.Context(user, fabsdk.WithOrg(orgName))

	resMgmt, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Creating resource management client failed")
	}

	instantiated, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok, err := isCCInstantiated(resMgmt, channelID, ccName, ccVersion)
			if err != nil {
				return &ok, err
			}
			if !ok && transientRetry {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Did NOT find instantiated chaincode [%s:%s] on one or more peers in [%s].", ccName, ccVersion, orgName), nil)
			}
			return &ok, nil
		},
	)

	if err != nil {
		s, ok := status.FromError(err)
		if ok && s.Code == status.GenericTransient.ToInt32() {
			return false, nil
		}
		return false, errors.WithMessage(err, "isCCInstantiated invocation failed")
	}

	return *instantiated.(*bool), nil
}

func isCCInstantiated(resMgmt *resmgmt.Client, channelID, ccName, ccVersion string) (bool, error) {
	chaincodeQueryResponse, err := resMgmt.QueryInstantiatedChaincodes(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return false, errors.WithMessage(err, "Query for instantiated chaincodes failed")
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == ccName && chaincode.Version == ccVersion {
			return true, nil
		}
	}
	return false, nil
}
