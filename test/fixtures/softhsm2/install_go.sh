#
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#!/bin/bash

# ----------------------------------------------------------------
# Install Golang
# ----------------------------------------------------------------

apt-get update
apt-get install -y -qq wget
mkdir -p $GOPATH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

cd /tmp
wget --quiet --no-check-certificate https://storage.googleapis.com/golang/go${GOVER}.linux-${ARCH}.tar.gz
tar -xvf go${GOVER}.linux-${ARCH}.tar.gz
mv go $GOROOT
chmod 775 $GOROOT