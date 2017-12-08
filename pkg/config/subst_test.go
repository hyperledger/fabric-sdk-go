/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestSubstOff(t *testing.T) {
	o := "foo${CRYPTOCONFIG_FIXTURES_PATH}${GOPATH}foo"
	s := substPathVars(o)
	e := "foo${CRYPTOCONFIG_FIXTURES_PATH}${GOPATH}foo"

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstCryptoConfigMiddle(t *testing.T) {
	o := "$foo${CRYPTOCONFIG_FIXTURES_PATH}foo"
	s := substPathVars(o)
	e := "$foo" + metadata.CryptoConfigPath + "foo"

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstCryptoConfigPrefix(t *testing.T) {
	o := "$foo${CRYPTOCONFIG_FIXTURES_PATH}"
	s := substPathVars(o)
	e := "$foo" + metadata.CryptoConfigPath

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}
func TestSubstCryptoConfigWithPostfix(t *testing.T) {
	o := "${CRYPTOCONFIG_FIXTURES_PATH}foo"
	s := substPathVars(o)
	e := metadata.CryptoConfigPath + "foo"

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstCryptoConfigOnly(t *testing.T) {
	o := "${CRYPTOCONFIG_FIXTURES_PATH}"
	s := substPathVars(o)
	e := metadata.CryptoConfigPath

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstAlmostVar1(t *testing.T) {
	o := "${FOO}"
	s := substPathVars(o)
	e := o

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstAlmostVar2(t *testing.T) {
	o := "${FOO${}${}$"
	s := substPathVars(o)
	e := o

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstNoVar(t *testing.T) {
	o := "foo"
	s := substPathVars(o)
	e := o

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstEmptyVar(t *testing.T) {
	o := ""
	s := substPathVars(o)
	e := o

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstGoPath1(t *testing.T) {
	o := "$foo${GOPATH}foo"
	s := substPathVars(o)
	e := "$foo" + goPath() + "foo"

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstGoPath2(t *testing.T) {
	o := "$foo${GOPATH}foo"
	s := substPathVars(o)
	e := "$foo" + goPath() + "foo"

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstGoPathAndVar(t *testing.T) {
	o := "$foo${GOPATH}foo${CRYPTOCONFIG_FIXTURES_PATH}"
	s := substPathVars(o)
	e := "$foo" + goPath() + "foo" + metadata.CryptoConfigPath

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}

func TestSubstGoPathOldStyle(t *testing.T) {
	o := "$foo$GOPATHfoo"
	s := substPathVars(o)
	e := "$foo" + goPath() + "foo"

	if s != e {
		t.Fatalf("Unexpected path substitution (%s, %s)", s, e)
	}
}
