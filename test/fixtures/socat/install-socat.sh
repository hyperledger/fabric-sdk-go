#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -xe

apt-get update

ARCH=`uname -m`
if [ $ARCH = "s390x" ]; then
  apt-get install -y perl
fi

apt-get install -y --no-install-recommends socat
