#
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

ARCH=`uname -m`

if [ $ARCH = "s390x" ]; then
  echo "deb http://ftp.us.debian.org/debian sid main" >> /etc/apt/sources.list
fi

apt-get update && \
apt-get -qq install -y make zlib1g-dev libbz2-dev libyaml-dev libltdl-dev libtool curl ca-certificates openssl git softhsm2 && \
mkdir -p /var/lib/softhsm/tokens/ && \
softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432 && \
rm -rf /var/lib/apt/lists/*