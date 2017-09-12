#!/bin/bash
#
# Copyright Greg Haskins, IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script runs Go linting and vetting tools

set -e

GO_CMD="${GO_CMD:-go}"
GOLINT_CMD=golint
GOFMT_CMD=gofmt
GOIMPORTS_CMD=goimports

PROJECT_PATH=$GOPATH/src/github.com/hyperledger/fabric-sdk-go

declare -a arr=(
"./api"
"./def"
"./pkg"
"./test"
)

echo "Running linters..."
for i in "${arr[@]}"
do
   echo "Checking $i"
   OUTPUT="$($GOLINT_CMD $i/...)"
   if [[ $OUTPUT ]]; then
      echo "You should check the following golint suggestions:"
      printf "$OUTPUT\n"
      echo "end golint suggestions"
   fi

   OUTPUT="$($GO_CMD vet $i/...)"
   if [[ $OUTPUT ]]; then
      echo "You should check the following govet suggestions:"
      printf "$OUTPUT\n"
      echo "end govet suggestions"
   fi

   found=`$GOFMT_CMD -l \`find $i -name "*.go" |grep -v "./vendor"\` 2>&1`
   if [ $? -ne 0 ]; then
      echo "The following files need reformatting with '$GO_FMT -w <file>':"
      printf "$badformat\n"
      exit 1
   fi

   OUTPUT="$($GOIMPORTS_CMD -srcdir $PROJECT_PATH -l $i)"
   if [[ $OUTPUT ]]; then
      echo "YOU MUST FIX THE FOLLOWING GOIMPORTS ERRORS:"
      printf "$OUTPUT\n"
      echo "END GOIMPORTS ERRORS"
      exit 1
   fi
done
