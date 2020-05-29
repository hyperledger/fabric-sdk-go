/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabricsdk enables Go developers to build solutions that interact with Hyperledger Fabric.
//
// Packages for end developer usage
//
// pkg/fabsdk: The main package of the Fabric SDK. This package enables creation of contexts based on
// configuration. These contexts are used by the client packages listed below.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/fabsdk
//
// pkg/client/channel: Provides channel transaction capabilities.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/channel
//
// pkg/client/event: Provides channel event capabilities.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/event
//
// pkg/client/ledger: Enables queries to a channel's underlying ledger.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/ledger
//
// pkg/client/resmgmt: Provides resource management capabilities such as installing chaincode.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt
//
// pkg/client/msp: Enables identity management capability.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/msp
//
// Basic workflow
//
//      1) Instantiate a fabsdk instance using a configuration.
//         Note: fabsdk maintains caches so you should minimize instances of fabsdk itself.
//      2) Create a context based on a user and organization, using your fabsdk instance.
//         Note: A channel context additionally requires the channel ID.
//      3) Create a client instance using its New func, passing the context.
//         Note: you create a new client instance for each context you need.
//      4) Use the funcs provided by each client to create your solution!
//      5) Call fabsdk.Close() to release resources and caches.
//
// Support for Hyperledger Fabric programming model
//
// In order to support the 'Gateway' programming model, the following package is provided:
//
// pkg/gateway: Enables Go developers to build client applications using the Hyperledger
// Fabric programming model as described in the 'Developing Applications' chapter of the Fabric
// documentation.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/gateway
//
package fabricsdk
