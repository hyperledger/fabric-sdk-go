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
    GO_VERSION=1.9.2 \
    PATH=$PATH:/opt/go/bin:/opt/gopath/bin \
    PKCS11_TOOL_URL=github.com/gbolo/go-util/p11tool

COPY install-softhsm2.sh /tmp
RUN bash /tmp/install-softhsm2.sh
CMD ["/bin/bash"]

