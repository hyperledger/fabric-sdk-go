/*
Copyright IBM Corp. 2017 All Rights Reserved.

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

package util

import (
	"fmt"
	"os"
)

const (
	defaultServerProtocol = "http"
	defaultServerAddr     = "localhost"
	defaultServerPort     = "7054"
)

// GetCommandLineOptValue searches the command line arguments for the
// specified option and returns the following value if found; otherwise
// it returns "".  If **remove** is true and it is found, the option
// and its value are removed from os.Args.
// For example, if command line is:
//    fabric-ca client enroll -config myconfig.json
// GetCommandLineOptValue("-config",true) returns "myconfig.json"
// and changes os.Args to
//    fabric-ca client enroll
func GetCommandLineOptValue(optName string, remove bool) string {
	for i := 0; i < len(os.Args)-1; i++ {
		if os.Args[i] == optName {
			val := os.Args[i+1]
			if remove {
				// Splice out the option and its value
				os.Args = append(os.Args[:i], os.Args[i+2:]...)
			}
			return val
		}
	}
	return ""
}

// GetServerURL returns the server's URL
func GetServerURL() string {
	return fmt.Sprintf("%s://%s:%s", GetServerProtocol(), GetServerAddr(), GetServerPort())
}

// GetServerProtocol returns the server's protocol
func GetServerProtocol() string {
	protocol := GetCommandLineOptValue("-protocol", false)
	if protocol != "" {
		return protocol
	}
	return defaultServerProtocol
}

// GetServerAddr returns the server's address
func GetServerAddr() string {
	addr := GetCommandLineOptValue("-address", false)
	if addr != "" {
		return addr
	}
	return defaultServerAddr
}

// GetServerPort returns the server's listening port
func GetServerPort() string {
	port := GetCommandLineOptValue("-port", false)
	if port != "" {
		return port
	}
	return defaultServerPort
}

// SetDefaultServerPort overrides the default CFSSL server port
// by adding the "-port" option to the command line if it was not
// already present.
func SetDefaultServerPort() {
	if len(os.Args) > 2 && GetCommandLineOptValue("-port", false) == "" {
		os.Args = append(os.Args, "-port", defaultServerPort)
	}
}
