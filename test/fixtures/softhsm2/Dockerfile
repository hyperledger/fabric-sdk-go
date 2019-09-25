#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BASE_GO_VERSION

FROM golang:${BASE_GO_VERSION}

COPY test/fixtures/softhsm2/install-softhsm2.sh /tmp
COPY scripts/_go/src/pkcs11helper /opt/workspace/pkcs11helper
RUN bash /tmp/install-softhsm2.sh
CMD ["/bin/bash"]

