/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// NumBlock is the number of block events received
type NumBlock uint

// NumChaincode is the number of chaincode events received
type NumChaincode uint

const (
	// ExpectOneBlock expects one block
	ExpectOneBlock NumBlock = 1
	// ExpectTwoBlocks expects two block
	ExpectTwoBlocks NumBlock = 2
	// ExpectThreeBlocks expects three block
	ExpectThreeBlocks NumBlock = 3
	// ExpectFourBlocks expects four block
	ExpectFourBlocks NumBlock = 4
	// ExpectFiveBlocks expects five block
	ExpectFiveBlocks NumBlock = 5
	// ExpectSixBlocks expects six block
	ExpectSixBlocks NumBlock = 6
	// ExpectSevenBlocks expects seven block
	ExpectSevenBlocks NumBlock = 7

	// ExpectOneCC expects one chaincode event
	ExpectOneCC NumChaincode = 1
	// ExpectTwoCC expects two chaincode event
	ExpectTwoCC NumChaincode = 2
	// ExpectThreeCC expects three chaincode event
	ExpectThreeCC NumChaincode = 3
	// ExpectFourCC expects four chaincode event
	ExpectFourCC NumChaincode = 4
)

// Received contains the number of block and chaincode events received
type Received struct {
	NumBlock     NumBlock
	NumChaincode NumChaincode
}

// NewReceived returns a new Received struct
func NewReceived(numBlock NumBlock, numChaincode NumChaincode) Received {
	return Received{NumBlock: numBlock, NumChaincode: numChaincode}
}
