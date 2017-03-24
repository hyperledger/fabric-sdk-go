/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


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

package fabricclient

import (
	"fmt"

	packager "github.com/hyperledger/fabric-sdk-go/fabric-client/packager"
)

// PackageCC ...
/**
 * Utility function to package a chaincode. The contents will be returned as a byte array.
 *
 * @param {string} chaincodePath required - String of the path to location of
 *                the source code of the chaincode
 * @param {string} chaincodeType optional - String of the type of chaincode
 *                 ['golang', 'car', 'java'] (default 'golang')
 * @returns {[]byte} byte array
 */
func PackageCC(chaincodePath string, chaincodeType string) ([]byte, error) {
	logger.Debugf("packager: chaincodePath: %s, chaincodeType: %s", chaincodePath, chaincodeType)
	if chaincodePath == "" {
		return nil, fmt.Errorf("Missing 'chaincodePath' parameter")
	}
	if chaincodeType == "" {
		chaincodeType = "golang"
	}
	logger.Debugf("packager: type %s ", chaincodeType)
	switch chaincodeType {
	case "golang":
		return packager.PackageGoLangCC(chaincodePath)
	}
	return nil, fmt.Errorf("Undefined 'chaincodeType' value")
}
