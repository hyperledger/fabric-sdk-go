#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
file_path=$1
exit_code=0
docker-compose --file=$file_path ps -q | xargs docker inspect -f '{{ .Name }},{{ .State.ExitCode }}' | \
while read name ; do
if echo "$name" | grep -q "softhsm2" 
then
    code="$(cut -d',' -f2 <<<"$name")"
    if (test "$code" -ne "$exit_code")
    then
        exit_code=1
    fi
fi
done
exit $exit_code

