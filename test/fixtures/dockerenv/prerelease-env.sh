#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This file contains environment overrides to enable testing
# against the latest pre-release target.
export FABRIC_FIXTURE_VERSION="v1.3"
export FABRIC_CRYPTOCONFIG_VERSION="v1"

export FABRIC_CA_FIXTURE_TAG="1.3.0-rc1"
export FABRIC_ORDERER_FIXTURE_TAG="1.3.0-rc1"
export FABRIC_PEER_FIXTURE_TAG="1.3.0-rc1"
export FABRIC_BUILDER_FIXTURE_TAG="1.3.0-rc1"

# override SDK configuration that loads crypto-config
export FABRIC_SDK_CLIENT_CRYPTOCONFIG_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
export FABRIC_SDK_CLIENT_ORDERERS_TLSCACERTS_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
export FABRIC_SDK_CLIENT_PEERS_PEER0_ORG1_EXAMPLE_COM_TLSCACERTS_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"
export FABRIC_SDK_CLIENT_PEERS_PEER0_ORG2_EXAMPLE_COM_TLSCACERTS_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem"