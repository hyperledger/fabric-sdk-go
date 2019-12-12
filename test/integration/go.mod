// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/hyperledger/fabric-sdk-go/test/integration

replace github.com/hyperledger/fabric-sdk-go => ../../

require (
	github.com/golang/protobuf v1.3.2
	github.com/hyperledger/fabric v1.4.4 // indirect
	github.com/hyperledger/fabric-protos-go v0.0.0-20191121202242-f5500d5e3e85
	github.com/hyperledger/fabric-sdk-go v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
	google.golang.org/grpc v1.23.0
)

go 1.13
