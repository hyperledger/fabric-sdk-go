/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"strings"
)

var configFile, configFileNoOrderer = fetchConfigFile()

// ConfigTestFile contains the path and filename of the config for integration tests
var ConfigTestFile = "../fixtures/config/" + configFile

//ConfigChBlockTestFile  the path and filename of the config for integration tests in which orderer config is not provided
var ConfigChBlockTestFile = "../fixtures/config/" + configFileNoOrderer

func fetchConfigFile() (string, string) {
	args := os.Args[1:]
	for _, arg := range args {
		if strings.Contains(arg, "configFile") {
			split := strings.Split(arg, "=")
			if split[1] != "" {
				if strings.Contains(split[1], "_local.yaml") {
					return split[1], "config_test_no_orderer_local.yaml"
				}
				return split[1], "config_test_no_orderer.yaml"
			}
		}
	}
	return "config_test.yaml", "config_test_no_orderer.yaml"
}
