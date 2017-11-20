#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

FILENAME=$1
NOTICE=$2
LICENSEID=$3

DEFAULT_LICENSEID="SPDX-License-Identifier: Apache-2.0"

if [ $# -lt 2 ]; then
    scriptname=`basename $0`
    echo "$scriptname FILENAME NOTICE LICENSEID"
    echo "FILENAME: Name of the source file to search"
    echo "NOTICE: Text to add below the license header"
    echo "LICENSEID: Text of license header to match - Default: ${DEFAULT_LICENSEID}"
    exit 1
fi

if [ -z "$3" ]; then
    LICENSEID="$DEFAULT_LICENSEID"
fi

# Find license identifier
if [ "$LICENSEID" != "::NONE::" ]; then
    licenseIdLine=`grep -Hn "${LICENSEID}" ${FILENAME} | head -n 1`

    # extract filename and license text line # 
    if [[ "$licenseIdLine" =~ ([^:]*):([^:]*):(.*) ]]; then
        licIDLineNum="${BASH_REMATCH[2]}"
    else
        echo "Unable to find license ID: ${LICENSEID}"
        exit 1
    fi
else
    licIDLineNum=1
fi

# find linecount
wcLines=`wc -l ${FILENAME}`
if [[ "$wcLines" =~ ([0-9]+) ]]; then
    lc="${BASH_REMATCH[1]}"
fi

# load portion of file after the license identifier
licenseIDPostText=`sed -n ${licIDLineNum},${lc}p ${FILENAME}`

# find closing comment
closingCommentLine=`echo "${licenseIDPostText}" | grep -Hn \*\/ | head -n 1`

# extract filename and license text line
newFile=""
if [ "$LICENSEID" != "::NONE::" ]; then
    if [[ "$closingCommentLine" =~ ([^:]*):([^:]*):(.*) ]]; then
        closingLineNum="${BASH_REMATCH[2]}"    
    else
        echo "Unable to find closing comment"
        exit 1
    fi

    # closingLineNum is in the context of the extracted licenseID post text (need to adjust the line #)
    closingLineNum=$(($closingLineNum+$licIDLineNum-1))

    # load existing file header
    newFile+=`sed -n 1,${closingLineNum}p ${FILENAME}`
    newFile+=$'\n'
else
    closingLineNum=0
fi

newFile+=$'/*\n'
newFile+=${NOTICE}
newFile+=$'\n*/\n'

# load existing rest of file
bodyLineNum=$(($closingLineNum+1))
newFile+=`sed -n ${bodyLineNum},${lc}p ${FILENAME}`

echo "$newFile"
