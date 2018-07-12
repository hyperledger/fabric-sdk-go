#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
ARG FABRIC_BASE_IMAGE=hyperledger/fabric-baseimage
ARG FABRIC_BASE_TAG

FROM ${FABRIC_BASE_IMAGE}:${FABRIC_BASE_TAG}

ENV GOPATH=/opt/gopath \
    GOROOT=/opt/go \
    PATH=$PATH:/opt/go/bin:/opt/gopath/bin

COPY test/fixtures/socat/install-socat.sh /tmp
RUN bash /tmp/install-socat.sh
ENTRYPOINT ["socat"]