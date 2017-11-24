/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gopackager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
	"testing"
)

// Test golang ChainCode packaging
func TestNewCCPackage(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}

	ccPackage, err := NewCCPackage("github.com", path.Join(pwd, "../../../../test/fixtures/testdata"))
	if err != nil {
		t.Fatalf("error from Create %v", err)
	}

	r := bytes.NewReader(ccPackage.Code)
	gzf, err := gzip.NewReader(r)
	if err != nil {
		t.Fatalf("error from gzip.NewReader %v", err)
	}
	tarReader := tar.NewReader(gzf)
	i := 0
	exampleccExist := false
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("error from tarReader.Next() %v", err)
		}

		if header.Name == "src/github.com/example_cc/example_cc.go" {
			exampleccExist = true
		}
		i++
	}

	if !exampleccExist {
		t.Fatalf("src/github.com/example_cc/example_cc.go not exist in tar file")
	}

}

// Test Package Go ChainCode
func TestEmptyCreate(t *testing.T) {

	_, err := NewCCPackage("", "")
	if err == nil {
		t.Fatalf("Package Empty GoLang CC must return an error.")
	}
}

// Test Bad Package Path for ChainCode packaging
func TestBadPackagePathGoLangCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}

	_, err = NewCCPackage("github.com", path.Join(pwd, "../../../../test/fixturesABC"))
	if err == nil {
		t.Fatalf("error expected from Create %v", err)
	}
}

// Test isSource set to true for any go readable files used in ChainCode packaging
func TestIsSourcePath(t *testing.T) {
	keep = []string{}
	isSrcVal := isSource("../")

	if isSrcVal {
		t.Fatalf("error expected when calling isSource %v", isSrcVal)
	}

	// reset keep
	keep = []string{".go", ".c", ".h"}
}

// Test packEntry and generateTarGz with empty file Descriptor
func TestEmptyPackEntry(t *testing.T) {
	emptyDescriptor := &Descriptor{"NewFile", ""}
	err := packEntry(nil, nil, emptyDescriptor)
	if err == nil {
		t.Fatal("packEntry call with empty descriptor info must throw an error")
	}

	_, err = generateTarGz([]*Descriptor{emptyDescriptor})
	if err == nil {
		t.Fatal("generateTarGz call with empty descriptor info must throw an error")
	}

}
