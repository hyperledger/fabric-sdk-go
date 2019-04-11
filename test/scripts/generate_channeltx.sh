#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

CONFIGTXGEN_CMD="${CONFIGTXGEN_CMD:-configtxgen}"
FIXTURES_PATH="${FIXTURES_PATH:-/opt/workspace/fabric-sdk-go/test/fixtures/}"
CHANNEL_DIR="${CHANNEL_DIR:-channel}"
CONFIG_DIR="${CONFIG_DIR:-config}"

if [ -z "$FABRIC_VERSION_DIR" ]; then
  echo "FABRIC_VERSION_DIR is required"
  exit 1
fi

declare -a oneOrgChannels=("mychannel")
declare -a twoOrgChannels=("orgchannel")
declare -a dsChannels=("dschannelsdk" "dschannelext")
declare -a orgs=("Org1MSP" "Org2MSP")

FIXTURES_CHANNEL_PATH=${FIXTURES_PATH}${FABRIC_VERSION_DIR}${CHANNEL_DIR}
export FABRIC_CFG_PATH=${FIXTURES_PATH}${FABRIC_VERSION_DIR}${CONFIG_DIR}

echo "Generating channel fixtures into ${FIXTURES_CHANNEL_PATH}"

mkdir -p ${FIXTURES_CHANNEL_PATH}

echo "Generating Orderer Genesis block"
$CONFIGTXGEN_CMD -profile TwoOrgsOrdererGenesis -outputBlock ${FIXTURES_CHANNEL_PATH}/twoorgs.genesis.block -channelID twoorgs

for i in "${oneOrgChannels[@]}"
do
   echo "Generating OneOrgChannel artifacts for channel: $i"

   echo "Generating channel configuration transaction"
   $CONFIGTXGEN_CMD -profile OneOrgChannel -outputCreateChannelTx .${FIXTURES_CHANNEL_PATH}/${i}.tx -channelID $i
done

for i in "${twoOrgChannels[@]}"
do
   echo "Generating TwoOrgsChannel artifacts for channel: $i"

   echo "Generating channel configuration transaction"
   $CONFIGTXGEN_CMD -profile TwoOrgsChannel -outputCreateChannelTx .${FIXTURES_CHANNEL_PATH}/${i}.tx -channelID $i

   for j in "${orgs[@]}"
   do
     echo "Generating anchor peer update for org $j"
     $CONFIGTXGEN_CMD -profile TwoOrgsChannel -outputAnchorPeersUpdate ${FIXTURES_CHANNEL_PATH}/${i}${j}anchors.tx -channelID $i -asOrg $j
   done
done

for i in "${dsChannels[@]}"
do
   echo "Generating DsChannel (Distributed Signing Identities Channel) artifacts for channel: $i"

   echo "Generating channel configuration transaction"
   $CONFIGTXGEN_CMD -profile DsChannel -outputCreateChannelTx .${FIXTURES_CHANNEL_PATH}/${i}.tx -channelID $i

   for j in "${orgs[@]}"
   do
     echo "Generating anchor peer update for org $j"
     $CONFIGTXGEN_CMD -profile DsChannel -outputAnchorPeersUpdate ${FIXTURES_CHANNEL_PATH}/${i}${j}anchors.tx -channelID $i -asOrg $j
   done
done