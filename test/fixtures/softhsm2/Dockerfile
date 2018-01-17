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
    PATH=$PATH:/opt/go/bin:/opt/gopath/bin

COPY test/fixtures/softhsm2/install-softhsm2.sh /tmp
COPY scripts/_go/src/pkcs11helper /opt/gopath/src/pkcs11helper
RUN bash /tmp/install-softhsm2.sh
CMD ["/bin/bash"]

