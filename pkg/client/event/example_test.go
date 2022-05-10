/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package event

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func Example() {

	ec, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, notifier, err := ec.RegisterChaincodeEvent("examplecc", "event123")
	if err != nil {
		fmt.Println("failed to register chaincode event")
	}
	defer ec.Unregister(registration)

	select {
	case ccEvent := <-notifier:
		fmt.Printf("received chaincode event %v\n", ccEvent)
	case <-time.After(time.Second * 5):
		fmt.Println("timeout while waiting for chaincode event")
	}

	// Timeout is expected since there is no event producer

	// Output: timeout while waiting for chaincode event

}

func ExampleNew() {

	ctx := mockChannelProvider("mychannel")

	ec, err := New(ctx, WithBlockEvents())
	if err != nil {
		fmt.Println(err)
	}

	if ec != nil {
		fmt.Println("event client created")
	} else {
		fmt.Println("event client is nil")
	}

	// Output: event client created

}

func ExampleClient_RegisterChaincodeEvent() {

	ec, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := ec.RegisterChaincodeEvent("examplecc", "event123")
	if err != nil {
		fmt.Println("failed to register chaincode event")
	}
	defer ec.Unregister(registration)

	fmt.Println("chaincode event registered successfully")

	// Output: chaincode event registered successfully

}

func ExampleClient_RegisterChaincodeEvent_NewService() {

	ec, err := New(mockChannelProvider("mychannel"), WithChaincodeID("examplecc"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := ec.RegisterChaincodeEvent("examplecc", "event123")
	if err != nil {
		fmt.Println("failed to register chaincode event")
	}
	defer ec.Unregister(registration)

	fmt.Println("chaincode event registered successfully")

	// Output: chaincode event registered successfully

}

func ExampleClient_RegisterChaincodeEvent_withPayload() {

	// If you require payload for chaincode events you have to use WithBlockEvents() option
	ec, err := New(mockChannelProvider("mychannel"), WithBlockEvents())
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := ec.RegisterChaincodeEvent("examplecc", "event123")
	if err != nil {
		fmt.Println("failed to register chaincode event")
	}
	defer ec.Unregister(registration)

	fmt.Println("chaincode event registered successfully")

	// Output: chaincode event registered successfully

}

func ExampleClient_RegisterTxStatusEvent() {

	ec, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := ec.RegisterTxStatusEvent("tx123")
	if err != nil {
		fmt.Println("failed to register tx status event")
	}
	defer ec.Unregister(registration)

	fmt.Println("tx status event registered successfully")

	// Output: tx status event registered successfully

}

func ExampleClient_RegisterBlockEvent() {

	ec, err := New(mockChannelProvider("mychannel"), WithBlockEvents())
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := ec.RegisterBlockEvent()
	if err != nil {
		fmt.Println("failed to register block event")
	}
	defer ec.Unregister(registration)

	fmt.Println("block event registered successfully")

	// Output: block event registered successfully

}

func ExampleClient_RegisterFilteredBlockEvent() {

	ec, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	registration, _, err := ec.RegisterFilteredBlockEvent()
	if err != nil {
		fmt.Println("failed to register filtered block event")
	}
	defer ec.Unregister(registration)

	fmt.Println("filtered block event registered successfully")

	// Output: filtered block event registered successfully

}

func mockChannelProvider(channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}
