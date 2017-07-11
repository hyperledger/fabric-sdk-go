/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fabrictxn "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn"
)

const (
	pollRetries = 5
)

// TestOrgsEndToEnd creates a channel with two organisations, installs chaincode
// on each of them, and finally invokes a transaction on an org2 peer and queries
// the result from an org1 peer
func TestOrgsEndToEnd(t *testing.T) {
	// Bootstrap network
	initializeFabricClient(t)
	loadOrgUsers(t)
	loadOrgPeers(t)
	loadOrderer(t)
	createTestChannel(t)
	joinTestChannel(t)
	installAndInstantiate(t)

	fmt.Printf("peer0 is %+v, peer1 is %+v\n", orgTestPeer0, orgTestPeer1)

	// Query initial value on org1 peer
	orgTestClient.SetUserContext(org1User)
	orgTestChannel.SetPrimaryPeer(orgTestPeer0)
	fcn := "invoke"
	result, err := fabrictxn.QueryChaincode(orgTestClient, orgTestChannel,
		"exampleCC", fcn, generateQueryArgs())
	failTestIfError(err, t)
	initialValue, err := strconv.Atoi(result)
	failTestIfError(err, t)

	// Change value on org2 peer
	orgTestClient.SetUserContext(org2User)
	orgTestChannel.SetPrimaryPeer(orgTestPeer1)
	_, err = fabrictxn.InvokeChaincode(orgTestClient, orgTestChannel, []apitxn.ProposalProcessor{orgTestPeer1},
		peer0EventHub, "exampleCC", fcn, generateInvokeArgs(), nil)
	failTestIfError(err, t)

	// Assert changed value on org1 peer
	var finalValue int
	for i := 0; i < pollRetries; i++ {
		finalValue = queryOrg1Peer(t)
		// If value has not propogated sleep with exponential backoff
		if initialValue+1 != finalValue {
			backoffFactor := math.Pow(2, float64(i))
			time.Sleep(time.Millisecond * 50 * time.Duration(backoffFactor))
		} else {
			break
		}
	}
	if initialValue+1 != finalValue {
		t.Fatalf("Org1 invoke result was not propagated to org2. Expected %d, got: %d",
			(initialValue + 1), finalValue)
	}
}

func queryOrg1Peer(t *testing.T) int {
	fcn := "invoke"

	orgTestClient.SetUserContext(org1User)
	orgTestChannel.SetPrimaryPeer(orgTestPeer0)
	result, err := fabrictxn.QueryChaincode(orgTestClient, orgTestChannel,
		"exampleCC", fcn, generateQueryArgs())
	failTestIfError(err, t)
	finalValue, err := strconv.Atoi(result)
	failTestIfError(err, t)

	return finalValue
}

func generateQueryArgs() []string {
	var args []string
	args = append(args, "query")
	args = append(args, "b")

	return args
}

func generateInvokeArgs() []string {
	var args []string
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	return args
}
