/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"
	"go/build"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

// goPath returns the current GOPATH. If the system
// has multiple GOPATHs then the first is used.
func goPath() string {
	gpDefault := build.Default.GOPATH
	gps := filepath.SplitList(gpDefault)

	return gps[0]
}

// substPathVars replaces instances of '${VARNAME}' (eg ${GOPATH}) with the variable.
// As a special case, $GOPATH is also replaced.
// NOTE: this function currently only performs substitution when the path string starts with $
//       as the path variables are intended to assist with testing.
func substPathVars(path string) string {
	if !strings.HasPrefix(path, "$") {
		return path
	}

	splits := strings.Split(path, "$")

	// Due to the first check above, the following code is currently not possible:
	//if len(splits) == 1 && path == splits[0] {
	//	// No variables are in the path
	//	return path
	//}

	var buffer bytes.Buffer
	buffer.WriteString(splits[0]) // first split precedes the first $ so should always be written
	for _, s := range splits[1:] {
		// special case for GOPATH
		if strings.HasPrefix(s, "GOPATH") {
			buffer.WriteString(goPath())
			buffer.WriteString(s[6:]) // Skip "GOPATH"
			continue
		}

		if !strings.HasPrefix(s, "{") {
			// not a variable
			buffer.WriteString("$")
			buffer.WriteString(s)
			continue
		}

		endPos := strings.Index(s, "}") // not worrying about embedded '{'
		if endPos == -1 {
			// not a variable
			buffer.WriteString("$")
			buffer.WriteString(s)
			continue
		}

		subs, ok := substVar(s[1:endPos]) // fix if not ASCII variable names
		if !ok {
			// not a variable
			buffer.WriteString("$")
			buffer.WriteString(s)
			continue
		}

		buffer.WriteString(subs)
		buffer.WriteString(s[endPos+1:]) // fix if not ASCII variable names
	}
	return buffer.String()
}

// substVar returns the substituted variable
func substVar(v string) (s string, ok bool) {
	// TODO: optimize if the number of variable names grows
	switch v {
	case "GOPATH":
		return goPath(), true
	case "CRYPTOCONFIG_FIXTURES_PATH":
		return metadata.CryptoConfigPath, true
	}
	return "", false
}
