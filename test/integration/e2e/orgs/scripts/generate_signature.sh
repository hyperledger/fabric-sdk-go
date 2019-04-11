#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

GO_CMD="${GO_CMD:-go}"
SCRIPT_DIR="$(dirname "$0")"
GOMOD_PATH=$(cd ${SCRIPT_DIR} && ${GO_CMD} env GOMOD)
PROJECT_MODULE=$(awk -F' ' '$1 == "module" {print $2}' ${GOMOD_PATH})
PROJECT_DIR=$(dirname ${GOMOD_PATH})

PWD_ORIG=$(pwd)
# need org in lowercase
ORG_N=$1
ORG=$(echo $ORG_N | tr '[:upper:]' '[:lower:]')
USER=$2
CH_CFG=$3
echo "CH_CFG is $CH_CFG, ORG is $ORG, USER is $USER"
KEY_PATH_DIR=${PROJECT_DIR}/../fixtures/fabric/v1/crypto-config/peerOrganizations/${ORG}.example.com/users/${USER}\@${ORG}.example.com/msp/keystore
SIGNATURE_PATH=$4
cd $KEY_PATH_DIR
KEY_NAME=$(ls)
echo "KEY_NAME is $KEY_NAME"
cd ${PWD_ORIG}
SIGNATURE_FILE=${SIGNATURE_PATH}/${CH_CFG}_${ORG_N}_${USER}_sbytes.txt.sha256
openssl dgst -sha256 -sign ${KEY_PATH_DIR}/${KEY_NAME} -out ${SIGNATURE_FILE} ${SIGNATURE_PATH}/${CH_CFG}_${ORG_N}_${USER}_sbytes.txt

echo "signature file generated name:[${CH_CFG}_${ORG_N}_${USER}_sbytes.txt.sha256] - content:[$(<${SIGNATURE_FILE})]"