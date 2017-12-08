#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This file contains environment overrides to enable testing
# against the latest pre-release target.
export FABRIC_FIXTURE_VERSION="v1.0"
export FABRIC_CRYPTOCONFIG_VERSION="v1"

export FABRIC_CA_FIXTURE_TAG="1.0.0"
export FABRIC_ORDERER_FIXTURE_TAG="1.0.0"
export FABRIC_PEER_FIXTURE_TAG="1.0.0"
export FABRIC_COUCHDB_FIXTURE_TAG="1.0.0"
export FABRIC_BUILDER_FIXTURE_TAG="1.0.0"
export FABRIC_BASEOS_FIXTURE_TAG="0.4.2"
export FABRIC_BASEIMAGE_FIXTURE_TAG="0.4.2"

# override configuration that loads crypto-config
FABRIC_SDK_CLIENT_CRYPTOCONFIG_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
FABRIC_SDK_CLIENT_ORDERERS_TLSCACERTS_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
FABRIC_SDK_CLIENT_PEERS_PEER0_ORG1_EXAMPLE_COM_TLSCACERTS_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"
FABRIC_SDK_CLIENT_PEERS_PEER0_ORG2_EXAMPLE_COM_TLSCACERTS_PATH='${GOPATH}'"/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem"
