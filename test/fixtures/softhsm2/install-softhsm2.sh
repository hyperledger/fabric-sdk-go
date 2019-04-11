#
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -xe

ARCH=`uname -m`

if [ $ARCH = "s390x" ]; then
  echo "deb http://ftp.us.debian.org/debian sid main" >> /etc/apt/sources.list
fi

apt-get update
apt-get install -y --no-install-recommends softhsm2 curl git gcc g++ libtool libltdl-dev
mkdir -p /var/lib/softhsm/tokens/
softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432

# Temporary fix for Fabric base image
unset GOCACHE

cd /opt/workspace/pkcs11helper/
go install pkcs11helper/cmd/pkcs11helper
