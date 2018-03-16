#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This file contains environment overrides to enable testing
# against a fabric service that doesn't require TLS client authentication.

export CORE_PEER_TLS_CLIENTAUTHREQUIRED=false
export CORE_PEER_TLS_CLIENTROOTCAS_FILES=""
export ORDERER_GENERAL_TLS_CLIENTAUTHENABLED=false
export ORDERER_GENERAL_TLS_CLIENTROOTCAS=""

export FABRIC_SDK_CLIENT_TLSCERTS_CLIENT_KEYFILE=""
export FABRIC_SDK_CLIENT_TLSCERTS_CLIENT_CERTFILE=""
