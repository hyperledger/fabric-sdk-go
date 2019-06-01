#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BASE_UBUNTU_VERSION

FROM ubuntu:${BASE_UBUNTU_VERSION}

ARG BASE_GO_VERSION

ENV GOPATH=/opt/gopath \
    GOROOT=/opt/go \
    PATH=$PATH:/opt/go/bin:/opt/gopath/bin \
    GOVER="${BASE_GO_VERSION}"

COPY test/fixtures/softhsm2/install_go.sh /tmp
RUN /tmp/install_go.sh

COPY test/fixtures/softhsm2/install-softhsm2.sh /tmp
COPY scripts/_go/src/pkcs11helper /opt/workspace/pkcs11helper
RUN bash /tmp/install-softhsm2.sh
CMD ["/bin/bash"]

