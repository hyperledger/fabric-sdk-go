/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"path/filepath"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	fabAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	javapackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/javapackager"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/nodepackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"github.com/pkg/errors"
)

var orgExpectedPeers = map[string]int{
	"Org1": 2,
	"Org2": 2,
}

const (
	defaultChannelID     = "mychannel"
	exampleCCName        = "example_cc"
	exampleCCPath        = "github.com/example_cc"
	exampleCCVersion     = "v0"
	examplePvtCCName     = "example_pvt_cc"
	examplePvtCCPath     = "github.com/example_pvt_cc"
	examplePvtCCVersion  = "v0"
	exampleUpgdPvtCCVer  = "v1"
	exampleJavaCCName    = "example_java_cc"
	exampleJavaCCPath    = "example_cc"
	exampleJavaCCVersion = "v0"
	exampleUpgdJavaCCVer = "v1"
	exampleNodeCCName    = "example_node_cc"
	exampleNodeCCPath    = "example_cc"
	exampleNodeCCVersion = "v0"
	exampleUpgdNodeCCVer = "v1"
)

// GenerateExamplePvtID supplies a chaincode name for example_pvt_cc
func GenerateExamplePvtID(randomize bool) string {
	suffix := "0"
	if randomize {
		suffix = GenerateRandomID()
	}

	return fmt.Sprintf("%s_fabtest_%s%s", examplePvtCCName, metadata.TestRunID, suffix)
}

// GenerateExampleID supplies a chaincode name for example_cc
func GenerateExampleID(randomize bool) string {
	suffix := "0"
	if randomize {
		suffix = GenerateRandomID()
	}

	return fmt.Sprintf("%s_fabtest_0%s%s", exampleCCName, metadata.TestRunID, suffix)
}

// GenerateExampleJavaID supplies a java chaincode name for example_cc
func GenerateExampleJavaID(randomize bool) string {
	suffix := "0"
	if randomize {
		suffix = GenerateRandomID()
	}

	return fmt.Sprintf("%s_fabtest_0%s%s", exampleJavaCCName, metadata.TestRunID, suffix)
}

// GenerateExampleNodeID supplies a node chaincode name for example_cc
func GenerateExampleNodeID(randomize bool) string {
	suffix := "0"
	if randomize {
		suffix = GenerateRandomID()
	}

	return fmt.Sprintf("%s_fabtest_0%s%s", exampleNodeCCName, metadata.TestRunID, suffix)
}

// PrepareExampleCC install and instantiate using resource management client
func PrepareExampleCC(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, chaincodeID string) error {
	const (
		channelID = defaultChannelID
	)

	instantiated, err := queryInstantiatedCCWithSDK(sdk, user, orgName, channelID, chaincodeID, exampleCCVersion, false)
	if err != nil {
		return errors.WithMessage(err, "Querying for instantiated status failed")
	}

	if instantiated {
		resetErr := resetExampleCC(sdk, user, orgName, channelID, chaincodeID, resetArgs)
		if resetErr != nil {
			return errors.WithMessage(resetErr, "Resetting example chaincode failed")
		}
		return nil
	}

	fmt.Printf("Installing and instantiating example chaincode...")
	start := time.Now()

	ccPolicy, err := prepareOneOrgPolicy(sdk, orgName)
	if err != nil {
		return errors.WithMessage(err, "CC policy could not be prepared")
	}

	orgContexts, err := prepareOrgContexts(sdk, user, []string{orgName})
	if err != nil {
		return errors.WithMessage(err, "Org contexts could not be prepared")
	}

	err = InstallExampleChaincode(orgContexts, chaincodeID)
	if err != nil {
		return errors.WithMessage(err, "Installing example chaincode failed")
	}

	err = InstantiateExampleChaincode(orgContexts, channelID, chaincodeID, ccPolicy)
	if err != nil {
		return errors.WithMessage(err, "Instantiating example chaincode failed")
	}

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Done [%d ms]\n", elapsed/time.Millisecond)

	return nil
}

// PrepareExampleCCLc install and instantiate using resource management client
func PrepareExampleCCLc(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, chaincodeID string) error {
	const (
		channelID = defaultChannelID
	)

	fmt.Printf("Installing and instantiating example chaincode..., cc = %v\n", chaincodeID)
	start := time.Now()

	ccPolicy, err := prepareOneOrgPolicy(sdk, orgName)
	if err != nil {
		return errors.WithMessage(err, "CC policy could not be prepared")
	}

	orgContexts, err := prepareOrgContexts(sdk, user, []string{orgName})
	if err != nil {
		return errors.WithMessage(err, "Org contexts could not be prepared")
	}

	packageID, err := InstallExampleChaincodeLc(orgContexts, chaincodeID, exampleCCVersion)
	if err != nil {
		return errors.WithMessage(err, "Installing example chaincode failed")
	}

	err = ApproveExampleChaincode(orgContexts, channelID, chaincodeID, exampleCCVersion, packageID, ccPolicy, 1)
	if err != nil {
		return errors.WithMessage(err, "Approve example chaincode failed")
	}

	err = QueryApprovedCC(orgContexts, chaincodeID, 1, channelID)
	if err != nil {
		return errors.WithMessage(err, "QueryApprovedCC example chaincode failed")
	}

	err = CheckCCCommitReadiness(orgContexts, chaincodeID, exampleCCVersion, 1, channelID, ccPolicy)
	if err != nil {
		return errors.WithMessage(err, "CheckCCCommitReadiness example chaincode failed")
	}

	err = CommitExampleChaincode(orgContexts, channelID, chaincodeID, exampleCCVersion, ccPolicy, 1)
	if err != nil {
		return errors.WithMessage(err, "Commit example chaincode failed")
	}

	err = QueryCommittedCC(orgContexts, chaincodeID, channelID, 1)
	if err != nil {
		return errors.WithMessage(err, "QueryCommittedCC example chaincode failed")
	}

	err = InitExampleChaincode(sdk, channelID, chaincodeID, orgContexts[0].OrgID)
	if err != nil {
		return errors.WithMessage(err, "Init example chaincode failed")
	}

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Done [%d ms]\n", elapsed/time.Millisecond)

	return nil
}

// InstallExampleChaincode installs the example chaincode to all peers in the given orgs
func InstallExampleChaincode(orgs []*OrgContext, ccID string) error {
	ccPkg, err := packager.NewCCPackage(exampleCCPath, GetDeployPath())
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, exampleCCPath, ccID, exampleCCVersion)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	return nil
}

// InstallExampleChaincodeLc installs the example chaincode to all peers in the given orgs
func InstallExampleChaincodeLc(orgs []*OrgContext, ccID, ccVersion string) (string, error) {
	label := ccID + "_" + ccVersion
	desc := &lcpackager.Descriptor{
		Path:  GetLcDeployPath(),
		Type:  pb.ChaincodeSpec_GOLANG,
		Label: label,
	}

	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		return "", errors.WithMessage(err, "creating chaincode package failed")
	}

	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}

	packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)

	for _, orgCtx := range orgs {
		_, err := orgCtx.ResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			return "", errors.WithMessage(err, "installing example chaincode failed")
		}
	}

	return packageID, nil
}

// ApproveExampleChaincode approve the example CC on the given channel
func ApproveExampleChaincode(orgs []*OrgContext, channelID, ccID, ccVersion, packageID, ccPolicyStr string, sequence int64, collConfigs ...*pb.CollectionConfig) error {
	ccPolicy, err := policydsl.FromString(ccPolicyStr)
	if err != nil {
		return errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}
	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              ccID,
		Version:           ccVersion,
		PackageID:         packageID,
		Sequence:          sequence,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      true,
		CollectionConfig:  collConfigs,
	}

	for _, orgCtx := range orgs {
		_, err := orgCtx.ResMgmt.LifecycleApproveCC(channelID, approveCCReq, resmgmt.WithTargets(orgCtx.Peers...), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			return errors.WithMessage(err, "approve example chaincode failed")
		}
	}

	return nil
}

// QueryApprovedCC query approve of the example CC on the given channel
func QueryApprovedCC(mc []*OrgContext, ccName string, sequence int64, channelID string) error {

	// Query approve cc
	queryApprovedCCReq := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     ccName,
		Sequence: sequence,
	}
	for _, orgCtx := range mc {
		for _, p := range orgCtx.Peers {
			err := queryApprovedCC(orgCtx, channelID, p, queryApprovedCCReq)
			if err != nil {
				return errors.WithMessage(err, "QueryApprovedCC example chaincode failed")
			}
		}
	}
	return nil

}

func queryApprovedCC(orgCtx *OrgContext, channelID string, p fabAPI.Peer, req resmgmt.LifecycleQueryApprovedCCRequest) error {

	_, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			resp1, err := orgCtx.ResMgmt.LifecycleQueryApprovedCC(channelID, req, resmgmt.WithTargets(p), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("QueryApprovedCC returned : %v", resp1), nil)
			}
			return resp1, err
		},
	)
	if err != nil {
		return errors.WithMessage(err, "QueryApprovedCC example chaincode failed")
	}

	return nil
}

// CheckCCCommitReadiness checkcommit the example CC on the given channel
func CheckCCCommitReadiness(mc []*OrgContext, ccName, ccVersion string, sequence int64, channelID string, ccPolicyStr string, collConfigs ...*pb.CollectionConfig) error {
	ccPolicy, err := policydsl.FromString(ccPolicyStr)
	if err != nil {
		return errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}
	req := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              ccName,
		Version:           ccVersion,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		Sequence:          sequence,
		InitRequired:      true,
		CollectionConfig:  collConfigs,
	}

	for _, orgCtx := range mc {
		for _, p := range orgCtx.Peers {
			err = checkCCCommitReadiness(orgCtx, channelID, p, req)
			if err != nil {
				return errors.WithMessage(err, "LifecycleCheckCCCommitReadiness example chaincode failed")
			}
		}
	}
	return nil
}

func checkCCCommitReadiness(orgCtx *OrgContext, channelID string, p fabAPI.Peer, req resmgmt.LifecycleCheckCCCommitReadinessRequest) error {

	_, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			resp1, err := orgCtx.ResMgmt.LifecycleCheckCCCommitReadiness(channelID, req, resmgmt.WithTargets(p))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("LifecycleCheckCCCommitReadiness returned : %v", resp1), nil)
			}
			flag := true
			for _, r := range resp1.Approvals {
				flag = flag && r
			}
			if !flag {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("LifecycleCheckCCCommitReadiness returned : %v", resp1), nil)
			}
			return resp1, err
		},
	)
	if err != nil {
		return errors.WithMessage(err, "LifecycleCheckCCCommitReadiness example chaincode failed")
	}

	return nil
}

// CommitExampleChaincode approve the example CC on the given channel
func CommitExampleChaincode(orgs []*OrgContext, channelID, ccID, ccVersion, ccPolicyStr string, sequence int64, collConfigs ...*pb.CollectionConfig) error {
	ccPolicy, err := policydsl.FromString(ccPolicyStr)
	if err != nil {
		return errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	req := resmgmt.LifecycleCommitCCRequest{
		Name:              ccID,
		Version:           ccVersion,
		Sequence:          sequence,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      true,
		CollectionConfig:  collConfigs,
	}

	_, err = orgs[0].ResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.WithMessage(err, "commit example chaincode failed")
	}

	return nil

}

// InitExampleChaincode init the example CC on the given channel
func InitExampleChaincode(sdk *fabsdk.FabricSDK, channelID, ccID string, orgName string) error {
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		return errors.WithMessage(err, "init example chaincode failed")
	}

	// init
	_, err = client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "init", Args: ExampleCCInitArgsLc(), IsInit: true},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		return errors.WithMessage(err, "init example chaincode failed")
	}
	return nil

}

// QueryCommittedCC query commit of the example CC on the given channel
func QueryCommittedCC(mc []*OrgContext, ccName string, channelID string, sequence int64) error {

	req := resmgmt.LifecycleQueryCommittedCCRequest{
		Name: ccName,
	}
	for _, orgCtx := range mc {
		for _, p := range orgCtx.Peers {
			err := queryCommittedCC(orgCtx, ccName, channelID, sequence, p, req)
			if err != nil {
				return errors.WithMessage(err, "queryCommittedCC example chaincode failed")
			}
		}
	}
	return nil

}

func queryCommittedCC(orgCtx *OrgContext, ccName string, channelID string, sequence int64, p fabAPI.Peer, req resmgmt.LifecycleQueryCommittedCCRequest) error {

	_, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			resp1, err := orgCtx.ResMgmt.LifecycleQueryCommittedCC(channelID, req, resmgmt.WithTargets(p))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("LifecycleQueryCommittedCC returned : %v", resp1), nil)
			}
			flag := false
			for _, r := range resp1 {
				if r.Name == ccName && r.Sequence == sequence {
					flag = true
					break
				}
			}
			if !flag {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("LifecycleQueryCommittedCC returned : %v", resp1), nil)
			}
			return resp1, err
		},
	)
	if err != nil {
		return errors.WithMessage(err, "queryCommittedCC example chaincode failed")
	}

	return nil

}

// InstallExamplePvtChaincode installs the example pvt chaincode to all peers in the given orgs
func InstallExamplePvtChaincode(orgs []*OrgContext, ccID string) error {
	ccPkg, err := packager.NewCCPackage(examplePvtCCPath, GetDeployPath())
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, examplePvtCCPath, ccID, examplePvtCCVersion)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	return nil
}

// InstallExamplePvtChaincodeLc installs the example chaincode to all peers in the given orgs
func InstallExamplePvtChaincodeLc(orgs []*OrgContext, ccID, ccVersion string) (string, error) {
	label := ccID + "_" + ccVersion
	desc := &lcpackager.Descriptor{
		Path:  GetLcPvtDeployPath(),
		Type:  pb.ChaincodeSpec_GOLANG,
		Label: label,
	}

	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		return "", errors.WithMessage(err, "creating chaincode package failed")
	}

	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}

	packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)

	for _, orgCtx := range orgs {
		_, err := orgCtx.ResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			return "", errors.WithMessage(err, "installing example chaincode failed")
		}
	}

	return packageID, nil
}

// InstallExampleJavaChaincode installs the example java chaincode to all peers in the given orgs
func InstallExampleJavaChaincode(orgs []*OrgContext, ccID string) error {
	ccPkg, err := javapackager.NewCCPackage(filepath.Join(GetJavaDeployPath(), exampleJavaCCPath))
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, exampleJavaCCPath, ccID, exampleJavaCCVersion)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	return nil
}

// InstallExampleNodeChaincode installs the example node chaincode to all peers in the given orgs
func InstallExampleNodeChaincode(orgs []*OrgContext, ccID string) error {
	ccPkg, err := nodepackager.NewCCPackage(filepath.Join(GetNodeDeployPath(), exampleNodeCCPath))
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, exampleNodeCCPath, ccID, exampleNodeCCVersion)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	return nil
}

// InstantiateExampleChaincodeLc install and instantiate using resource management client
func InstantiateExampleChaincodeLc(sdk *fabsdk.FabricSDK, orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	start := time.Now()
	packageID, err := InstallExampleChaincodeLc(orgs, ccID, exampleCCVersion)
	if err != nil {
		return errors.WithMessage(err, "Installing example chaincode failed")
	}

	err = instantiateExampleChaincodeLc(sdk, orgs, channelID, ccID, exampleCCVersion, ccPolicy, packageID, 1, collConfigs...)
	if err != nil {
		return errors.WithMessage(err, "init example chaincode failed")
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Done [%d ms]\n", elapsed/time.Millisecond)

	return nil
}

// InstantiatePvtExampleChaincodeLc install and instantiate using resource management client
func InstantiatePvtExampleChaincodeLc(sdk *fabsdk.FabricSDK, orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	start := time.Now()
	packageID, err := InstallExamplePvtChaincodeLc(orgs, ccID, exampleCCVersion)
	if err != nil {
		return errors.WithMessage(err, "Installing example chaincode failed")
	}

	err = instantiateExampleChaincodeLc(sdk, orgs, channelID, ccID, exampleCCVersion, ccPolicy, packageID, 1, collConfigs...)
	if err != nil {
		return errors.WithMessage(err, "init example chaincode failed")
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Done [%d ms]\n", elapsed/time.Millisecond)

	return nil
}

func instantiateExampleChaincodeLc(sdk *fabsdk.FabricSDK, orgs []*OrgContext, channelID, ccID, ccVersion, ccPolicy, packageID string, sequence int64, collConfigs ...*pb.CollectionConfig) error {

	err := ApproveExampleChaincode(orgs, channelID, ccID, ccVersion, packageID, ccPolicy, sequence, collConfigs...)
	if err != nil {
		return errors.WithMessage(err, "Approve example chaincode failed")
	}

	err = QueryApprovedCC(orgs, ccID, sequence, channelID)
	if err != nil {
		return errors.WithMessage(err, "QueryApprovedCC example chaincode failed")
	}

	err = CheckCCCommitReadiness(orgs, ccID, ccVersion, sequence, channelID, ccPolicy, collConfigs...)
	if err != nil {
		return errors.WithMessage(err, "CheckCCCommitReadiness example chaincode failed")
	}

	err = CommitExampleChaincode(orgs, channelID, ccID, ccVersion, ccPolicy, sequence, collConfigs...)
	if err != nil {
		return errors.WithMessage(err, "Commit example chaincode failed")
	}

	err = QueryCommittedCC(orgs, ccID, channelID, sequence)
	if err != nil {
		return errors.WithMessage(err, "QueryCommittedCC example chaincode failed")
	}

	err = InitExampleChaincode(sdk, channelID, ccID, orgs[0].OrgID)
	if err != nil {
		return errors.WithMessage(err, "Init example chaincode failed")
	}

	return nil
}

// UpgradeExamplePvtChaincodeLc upgrades the instantiated example pvt CC on the given channel
func UpgradeExamplePvtChaincodeLc(sdk *fabsdk.FabricSDK, orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	start := time.Now()
	packageID, err := InstallExamplePvtChaincodeLc(orgs, ccID, exampleUpgdPvtCCVer)
	if err != nil {
		return errors.WithMessage(err, "Installing example chaincode failed")
	}

	err = instantiateExampleChaincodeLc(sdk, orgs, channelID, ccID, exampleUpgdPvtCCVer, ccPolicy, packageID, 2, collConfigs...)
	if err != nil {
		return errors.WithMessage(err, "init example chaincode failed")
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Done [%d ms]\n", elapsed/time.Millisecond)

	return nil
}

// InstantiateExampleChaincode instantiates the example CC on the given channel
func InstantiateExampleChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	_, err := InstantiateChaincode(orgs[0].ResMgmt, channelID, ccID, exampleCCPath, exampleCCVersion, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
}

// InstantiateExamplePvtChaincode instantiates the example pvt CC on the given channel
func InstantiateExamplePvtChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	_, err := InstantiateChaincode(orgs[0].ResMgmt, channelID, ccID, examplePvtCCPath, examplePvtCCVersion, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
}

// InstantiateExampleJavaChaincode instantiates the example CC on the given channel
func InstantiateExampleJavaChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	_, err := InstantiateJavaChaincode(orgs[0].ResMgmt, channelID, ccID, exampleJavaCCPath, exampleJavaCCVersion, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
}

// InstantiateExampleNodeChaincode instantiates the example CC on the given channel
func InstantiateExampleNodeChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	_, err := InstantiateNodeChaincode(orgs[0].ResMgmt, channelID, ccID, exampleNodeCCPath, exampleNodeCCVersion, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
}

// UpgradeExamplePvtChaincode upgrades the instantiated example pvt CC on the given channel
func UpgradeExamplePvtChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	// first install the CC with the upgraded cc version
	ccPkg, err := packager.NewCCPackage(examplePvtCCPath, GetDeployPath())
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}
	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, examplePvtCCPath, ccID, exampleUpgdPvtCCVer)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	// now upgrade cc
	_, err = UpgradeChaincode(orgs[0].ResMgmt, channelID, ccID, examplePvtCCPath, exampleUpgdPvtCCVer, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
}

// UpgradeExampleJavaChaincode upgrades the instantiated example java CC on the given channel
func UpgradeExampleJavaChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	ccPkg, err := javapackager.NewCCPackage(filepath.Join(GetJavaDeployPath(), exampleJavaCCPath))
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, exampleJavaCCPath, ccID, exampleUpgdJavaCCVer)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	// now upgrade cc
	_, err = UpgradeJavaChaincode(orgs[0].ResMgmt, channelID, ccID, exampleJavaCCPath, exampleUpgdJavaCCVer, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
}

// UpgradeExampleNodeChaincode upgrades the instantiated example java CC on the given channel
func UpgradeExampleNodeChaincode(orgs []*OrgContext, channelID, ccID, ccPolicy string, collConfigs ...*pb.CollectionConfig) error {
	ccPkg, err := nodepackager.NewCCPackage(filepath.Join(GetJavaDeployPath(), exampleNodeCCPath))
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	err = InstallChaincodeWithOrgContexts(orgs, ccPkg, exampleNodeCCPath, ccID, exampleUpgdNodeCCVer)
	if err != nil {
		return errors.WithMessage(err, "installing example chaincode failed")
	}

	// now upgrade cc
	_, err = UpgradeNodeChaincode(orgs[0].ResMgmt, channelID, ccID, exampleNodeCCPath, exampleUpgdNodeCCVer, ccPolicy, ExampleCCInitArgs(), collConfigs...)
	return err
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

func prepareOrgContexts(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgNames []string) ([]*OrgContext, error) {
	orgContexts := make([]*OrgContext, len(orgNames))

	for i, orgName := range orgNames {
		clientContext := sdk.Context(user, fabsdk.WithOrg(orgName))

		resMgmt, err := resmgmt.New(clientContext)
		if err != nil {
			return nil, errors.WithMessage(err, "Creating resource management client failed")
		}

		expectedPeers, ok := orgExpectedPeers[orgName]
		if !ok {
			return nil, errors.WithMessage(err, "unknown org name")
		}
		peers, err := DiscoverLocalPeers(clientContext, expectedPeers)
		if err != nil {
			return nil, errors.WithMessage(err, "local peers could not be determined")
		}

		orgCtx := OrgContext{
			OrgID:       orgName,
			CtxProvider: clientContext,
			ResMgmt:     resMgmt,
			Peers:       peers,
		}
		orgContexts[i] = &orgCtx
	}
	return orgContexts, nil
}

func prepareOneOrgPolicy(sdk *fabsdk.FabricSDK, orgName string) (string, error) {
	mspID, err := orgMSPID(sdk, orgName)
	if err != nil {
		return "", errors.WithMessage(err, "MSP ID could not be determined")
	}

	return fmt.Sprintf("AND('%s.member')", mspID), nil
}

func orgMSPID(sdk *fabsdk.FabricSDK, orgName string) (string, error) {
	configBackend, err := sdk.Config()
	if err != nil {
		return "", errors.WithMessage(err, "failed to get config backend")
	}

	endpointConfig, err := fab.ConfigFromBackend(configBackend)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get endpoint config")
	}

	mspID, ok := comm.MSPID(endpointConfig, orgName)
	if !ok {
		return "", errors.New("looking up MSP ID failed")
	}

	return mspID, nil
}

func queryInstantiatedCCWithSDK(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, channelID, ccName, ccVersion string, transientRetry bool) (bool, error) {
	clientContext := sdk.Context(user, fabsdk.WithOrg(orgName))

	resMgmt, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Creating resource management client failed")
	}

	return queryInstantiatedCC(resMgmt, orgName, channelID, ccName, ccVersion, transientRetry)
}

func queryInstantiatedCC(resMgmt *resmgmt.Client, orgName string, channelID, ccName, ccVersion string, transientRetry bool) (bool, error) {

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
