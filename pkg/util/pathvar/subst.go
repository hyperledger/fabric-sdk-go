/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pathvar

import (
	"bytes"
	"go/build"
	"os"
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

// Subst replaces instances of '${VARNAME}' (eg ${GOPATH}) with the variable.
// Variables names that are not set by the SDK are replaced with the environment variable.
func Subst(path string) string {
	const (
		sepPrefix = "${"
		sepSuffix = "}"
	)

	splits := strings.Split(path, sepPrefix)

	var buffer bytes.Buffer

	// first split precedes the first sepPrefix so should always be written
	buffer.WriteString(splits[0]) // nolint: gas

	for _, s := range splits[1:] {
		subst, rest := substVar(s, sepPrefix, sepSuffix)
		buffer.WriteString(subst) // nolint: gas
		buffer.WriteString(rest)  // nolint: gas
	}

	return buffer.String()
}

// substVar searches for an instance of a variables name and replaces them with their value.
// The first return value is substituted portion of the string or noMatch if no replacement occurred.
// The second return value is the unconsumed portion of s.
func substVar(s string, noMatch string, sep string) (string, string) {
	endPos := strings.Index(s, sep)
	if endPos == -1 {
		return noMatch, s
	}

	v, ok := lookupVar(s[:endPos])
	if !ok {
		return noMatch, s
	}

	return v, s[endPos+1:]
}

// lookupVar returns the value of the variable.
// The local variable table is consulted first, followed by environment variables.
// Returns false if the variable doesn't exist.
func lookupVar(v string) (string, bool) {
	// TODO: optimize if the number of variable names grows
	switch v {
	case "FABRIC_SDK_GO_PROJECT_PATH":
		return metadata.GetProjectPath(), true
	case "GOPATH":
		return goPath(), true
	case "CRYPTOCONFIG_FIXTURES_PATH":
		return metadata.CryptoConfigPath, true
	}
	return os.LookupEnv(v)
}
