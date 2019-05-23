#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BASE_UBUNTU_VERSION

FROM ubuntu:${BASE_UBUNTU_VERSION}

ENV GOPATH=/opt/gopath \
    GOROOT=/opt/go \
    PATH=$PATH:/opt/go/bin:/opt/gopath/bin

COPY test/fixtures/socat/install-socat.sh /tmp
RUN bash /tmp/install-socat.sh
ENTRYPOINT ["socat"]