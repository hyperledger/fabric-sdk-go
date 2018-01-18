/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package util

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	factory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/cryptosuitebridge"
)

type alg struct {
	hashFun func([]byte) string
}

const defaultAlg = "sha256"

var availableIDgenAlgs = map[string]alg{
	defaultAlg: {GenerateIDfromTxSHAHash},
}

// ComputeSHA256 returns SHA2-256 on data
func ComputeSHA256(data []byte) (hash []byte) {
	hash, err := factory.GetDefault().Hash(data, factory.GetSHA256Opts())
	if err != nil {
		panic(fmt.Errorf("Failed computing SHA256 on [% x]", data))
	}
	return
}

// CreateUtcTimestamp returns a google/protobuf/Timestamp in UTC
func CreateUtcTimestamp() *timestamp.Timestamp {
	now := time.Now().UTC()
	secs := now.Unix()
	nanos := int32(now.UnixNano() - (secs * 1000000000))
	return &(timestamp.Timestamp{Seconds: secs, Nanos: nanos})
}

// GenerateIDfromTxSHAHash generates SHA256 hash using Tx payload
func GenerateIDfromTxSHAHash(payload []byte) string {
	return fmt.Sprintf("%x", ComputeSHA256(payload))
}

const testchainid = "testchainid"
const testorgid = "**TEST_ORGID**"

// ConcatenateBytes is useful for combining multiple arrays of bytes, especially for
// signatures or digests over multiple fields
func ConcatenateBytes(data ...[]byte) []byte {
	finalLength := 0
	for _, slice := range data {
		finalLength += len(slice)
	}
	result := make([]byte, finalLength)
	last := 0
	for _, slice := range data {
		for i := range slice {
			result[i+last] = slice[i]
		}
		last += len(slice)
	}
	return result
}

const DELIMITER = "."
