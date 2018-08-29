/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	reqContext "context"

	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

//TestParentContext tests to make sure external grpc context can be passed as a parent context to highlevel functions
func TestParentContext(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	target := mainTestSetup.Targets[0]
	chaincodeID := mainChaincodeID

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))
	org1AdminChannelContext := sdk.ChannelContext(mainTestSetup.ChannelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	//prepare context
	ctx, err := org1AdminClientContext()
	if err != nil {
		t.Fatal("failed to get client context")
	}

	//get parent context and cancel
	parentContext, cancel := context.NewRequest(ctx, context.WithTimeout(20*time.Second))
	//Cancel in advance - to make sure test fails with 'context cancelled' error
	cancel()

	// Resource management client
	resClient, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	_, err = resClient.QueryChannels(resmgmt.WithTargetEndpoints(target), resmgmt.WithParentContext(parentContext), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context cancelled error but got: %s", err)
	}

	// Channel client
	chClient, err := channel.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	_, err = chClient.Query(channel.Request{ChaincodeID: chaincodeID, Fcn: "invoke", Args: integration.ExampleCCDefaultQueryArgs()},
		channel.WithParentContext(parentContext),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err == nil || !strings.Contains(err.Error(), "request timed out or been cancelled") {
		t.Fatalf("expected context cancelled error but got: %s", err)
	}

	// ledger client
	testLedgerClient(org1AdminChannelContext, t, parentContext)

}

func testLedgerClient(org1AdminChannelContext contextApi.ChannelProvider, t *testing.T, parentContext reqContext.Context) {
	legerClient, err := ledger.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}
	_, err = legerClient.QueryInfo(ledger.WithParentContext(parentContext))
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context cancelled error but got: %s", err)
	}
}
