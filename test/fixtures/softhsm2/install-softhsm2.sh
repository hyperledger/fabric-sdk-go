#
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -xe

apt-get update

ARCH=`uname -m`
if [ $ARCH = "s390x" ]; then
  apt-get install -y perl
fi

apt-get install -y --no-install-recommends softhsm2 curl git gcc g++ libtool libltdl-dev
mkdir -p /var/lib/softhsm/tokens/
softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432

cd /opt/workspace/pkcs11helper/
go install pkcs11helper/cmd/pkcs11helper
