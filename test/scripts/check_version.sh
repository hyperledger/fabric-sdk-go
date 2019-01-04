#!/bin/bash
# 
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Checking Go version"
go version

GO_VER=`grep GO_VER ci.properties | awk -F "=" '{print $2}'`
GO_MIN_VER=`grep GO_MIN_VER ci.properties | awk -F "=" '{print $2}'`
GO_MAX_VER=`grep GO_MAX_VER ci.properties | awk -F "=" '{print $2}'`

function isGoVersionValid {
    # Check GO_MIN_VER, it must exist in ci.properties and be valid
    # GO_VER must be >= GO_MIN_VER
    # GO_MAX_VER must be >= GO_MIN_VER
    GO_MIN_VER_MAJOR=`echo $GO_MIN_VER | awk -F "." '{print $1}'`
    if [ -z $GO_MIN_VER_MAJOR ] || [ $GO_MIN_VER_MAJOR = "" ]; then
        echo "ERROR: GO_MIN_VER is not specified in ci.properties properly"
        exit 1
    fi

    GO_MIN_VER_MINOR=`echo $GO_MIN_VER | awk -F "." '{print $2}'`
    if [ -z $GO_MIN_VER_MINOR ] || [ $GO_MIN_VER_MINOR = "" ]; then
        echo "ERROR: GO_MIN_VER is not specified in ci.properties properly"
        exit 1
    fi
    GO_MIN_VER_RELEASE=`echo $GO_MIN_VER | awk -F "." '{print $3}'`
    if [ -z $GO_MIN_VER_RELEASE ]; then
        GO_MIN_VER_RELEASE=0
    fi

    # Check GO_MAX_VER, it must exist in ci.properties and be valid
    # GO_VER must be <= GO_MAX_VER
    # GO_MAX_VER must be >= GO_MIN_VER
    GO_MAX_VER_MAJOR=`echo $GO_MAX_VER | awk -F "." '{print $1}'`
    if [ -z $GO_MAX_VER_MAJOR ] || [ $GO_MAX_VER_MAJOR = "" ]; then
        echo "ERROR: GO_MAX_VER is not specified in ci.properties properly"
        exit 1
    fi
    if [ $GO_MAX_VER_MAJOR -lt $GO_MIN_VER_MAJOR ]; then
        echo "ERROR: GO_MAX_VER (${GO_MAX_VER}) is smaller then GO_MIN_VER (${GO_MIN_VER}) in ci.properties"
        exit 1
    fi

    GO_MAX_VER_MINOR=`echo $GO_MAX_VER | awk -F "." '{print $2}'`
    if [ -z $GO_MAX_VER_MINOR ] || [ $GO_MAX_VER_MINOR = "" ]; then
        echo "ERROR: GO_MAX_VER is not specified in ci.properties properly"
        exit 1
    fi
    if [ $GO_MAX_VER_MAJOR -eq $GO_MIN_VER_MAJOR ] && [ $GO_MAX_VER_MINOR -lt $GO_MIN_VER_MINOR ]; then
        echo "ERROR: GO_MAX_VER (${GO_MAX_VER}) is smaller then GO_MIN_VER (${GO_MIN_VER}) in ci.properties"
        exit 1
    fi

    GO_MAX_VER_RELEASE=`echo $GO_MAX_VER | awk -F "." '{print $3}'`
    if [ -z $GO_MAX_VER_RELEASE ]; then
        GO_MAX_VER_RELEASE=0
    fi
    if [ $GO_MAX_VER_MAJOR -eq $GO_MIN_VER_MAJOR ] && [ $GO_MAX_VER_MINOR -eq $GO_MIN_VER_MINOR ]; then
        if [ $GO_MAX_VER_RELEASE -lt $GO_MIN_VER_RELEASE ]; then
            echo "ERROR: GO_MAX_VER (${GO_MAX_VER}) is smaller then GO_MIN_VER (${GO_MIN_VER}) in ci.properties"
            exit 1
        fi
    fi

    GO_MAJOR_VERSION=`go version |awk '{print $3}' | awk -F "." '{print substr($1,3)}'`
    if [ $GO_MAJOR_VERSION -lt $GO_MIN_VER_MAJOR ] || [ $GO_MAJOR_VERSION -gt $GO_MAX_VER_MAJOR ]; then
        return 1
    fi

    GO_MINOR_VERSION=`go version |awk '{print $3}' | awk -F "." '{print $2}'`
    if [ $GO_MAJOR_VERSION -eq $GO_MIN_VER_MAJOR ] && [ $GO_MINOR_VERSION -lt $GO_MIN_VER_MINOR ]; then
        return 1
    fi
    if [ $GO_MAJOR_VERSION -eq $GO_MAX_VER_MAJOR ] && [ $GO_MINOR_VERSION -gt $GO_MAX_VER_MINOR ]; then
        return 1
    fi

    GO_RELEASE_NO=`go version |awk '{print $3}' | awk -F "." '{print $3}'`
    if [ -z $GO_RELEASE_NO ]; then
        GO_RELEASE_NO=0
    fi
    if [ $GO_MAJOR_VERSION -eq $GO_MIN_VER_MAJOR ] && [ $GO_MINOR_VERSION -eq $GO_MIN_VER_MINOR ]; then
        if [ $GO_RELEASE_NO -lt $GO_MIN_VER_RELEASE ]; then
            return 1
        fi
    fi
    if [ $GO_MAJOR_VERSION -eq $GO_MAX_VER_MAJOR ] && [ $GO_MINOR_VERSION -eq $GO_MAX_VER_MINOR ]; then
        if [ $GO_RELEASE_NO -gt $GO_MAX_VER_RELEASE ]; then
            return 1
        fi
    fi

    return 0
}

if ! isGoVersionValid; then
    echo "You should install go ${GO_MIN_VER} to ${GO_MAX_VER} to run hyperledger fabric sdk tests"
    exit 1
fi
