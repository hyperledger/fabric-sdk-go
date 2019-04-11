/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pathvar

import (
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"
)

func TestSubstCryptoConfigMiddle(t *testing.T) {
	o := "$foo${CRYPTOCONFIG_FIXTURES_PATH}foo"
	s := Subst(o)
	e := "$foo" + metadata.CryptoConfigPath + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstCryptoConfigPrefix(t *testing.T) {
	o := "$foo${CRYPTOCONFIG_FIXTURES_PATH}"
	s := Subst(o)
	e := "$foo" + metadata.CryptoConfigPath

	assert.Equal(t, e, s, "Unexpected path substitution")
}
func TestSubstCryptoConfigWithPostfix(t *testing.T) {
	o := "${CRYPTOCONFIG_FIXTURES_PATH}foo"
	s := Subst(o)
	e := metadata.CryptoConfigPath + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstCryptoConfigOnly(t *testing.T) {
	o := "${CRYPTOCONFIG_FIXTURES_PATH}"
	s := Subst(o)
	e := metadata.CryptoConfigPath

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstNotAKey(t *testing.T) {
	const envKey = "FABGOSDK_TESTVAR"

	o := "${" + envKey + "}"
	s := Subst(o)
	e := o

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstAlmostVar(t *testing.T) {
	const envKey = "FABGOSDK_TESTVAR"

	o := "${" + envKey + "{}${}$"
	s := Subst(o)
	e := o

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstNoVar(t *testing.T) {
	o := "foo"
	s := Subst(o)
	e := o

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstEmptyVar(t *testing.T) {
	o := ""
	s := Subst(o)
	e := o

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstGoPath1(t *testing.T) {
	o := "$foo${GOPATH}foo"
	s := Subst(o)
	e := "$foo" + goPath() + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstGoPath2(t *testing.T) {
	o := "$foo${GOPATH}foo"
	s := Subst(o)
	e := "$foo" + goPath() + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstGoPathAndVar(t *testing.T) {
	o := "$foo${GOPATH}foo${CRYPTOCONFIG_FIXTURES_PATH}"
	s := Subst(o)
	e := "$foo" + goPath() + "foo" + metadata.CryptoConfigPath

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstGoPathOldStyle(t *testing.T) {
	o := "$foo$GOPATHfoo"
	s := Subst(o)
	e := "$foo$GOPATHfoo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestEnvSubst(t *testing.T) {

	const envKey = "FABGOSDK_TESTVAR"
	const envVar = "I AM SET"

	err := os.Setenv(envKey, envVar)
	assert.Nil(t, err, "Setenv should have succeeded")

	o := "$foo${" + envKey + "}foo"
	s := Subst(o)
	e := "$foo" + envVar + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")

	err = os.Unsetenv("FABGOSDK_TESTVAR")
	assert.Nil(t, err, "Unsetenv should have succeeded")
}

func TestEnvPriority(t *testing.T) {

	const envKey = "CRYPTOCONFIG_FIXTURES_PATH"
	const envVar = "I AM SET"

	err := os.Setenv(envKey, envVar)
	assert.Nil(t, err, "Setenv should have succeeded")

	o := "$foo${GOPATH}foo${CRYPTOCONFIG_FIXTURES_PATH}"
	s := Subst(o)
	e := "$foo" + goPath() + "foo" + metadata.CryptoConfigPath

	assert.Equal(t, e, s, "Unexpected path substitution (SDK variable should take priority)")

	err = os.Unsetenv("FABGOSDK_TESTVAR")
	assert.Nil(t, err, "Unsetenv should have succeeded")
}

func TestSubstMiddleNoVarPrefix(t *testing.T) {
	o := "foo${CRYPTOCONFIG_FIXTURES_PATH}${GOPATH}foo"
	s := Subst(o)
	e := "foo" + metadata.CryptoConfigPath + goPath() + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}

func TestSubstProjectPath(t *testing.T) {
	o := "$foo${FABRIC_SDK_GO_PROJECT_PATH}foo"
	s := Subst(o)
	e := "$foo" + metadata.GetProjectPath() + "foo"

	assert.Equal(t, e, s, "Unexpected path substitution")
}
