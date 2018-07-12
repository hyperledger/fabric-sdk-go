#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -xe

ARCH=`uname -m`

if [ $ARCH = "s390x" ]; then
  echo "deb http://ftp.us.debian.org/debian sid main" >> /etc/apt/sources.list
fi

apt-get update
apt-get install -y --no-install-recommends socat