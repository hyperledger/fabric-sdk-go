/*
 Copyright Mioto Yaku All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package nodepackager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test golang ChainCode packaging
func TestNewCCPackage(t *testing.T) {
	pwd, err := os.Getwd()
	assert.Nil(t, err, "error from os.Getwd %s", err)

	ccPackage, err := NewCCPackage(filepath.Join(pwd, "testdata"))
	assert.Nil(t, err, "error from Create %s", err)

	r := bytes.NewReader(ccPackage.Code)

	gzf, err := gzip.NewReader(r)
	assert.Nil(t, err, "error from gzip.NewReader %s", err)

	tarReader := tar.NewReader(gzf)
	i := 0
	var exampleccExist, eventMetaInfExists, examplecc1MetaInfExists, fooMetaInfoExists, metaInfFooExists bool
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		assert.Nil(t, err, "error from tarReader.Next() %s", err)

		exampleccExist = exampleccExist || header.Name == "src/example_cc/chaincode_example02.js"
		eventMetaInfExists = eventMetaInfExists || header.Name == "META-INF/sample-json/event.json"
		examplecc1MetaInfExists = examplecc1MetaInfExists || header.Name == "META-INF/example1.json"
		fooMetaInfoExists = fooMetaInfoExists || strings.HasPrefix(header.Name, "foo-META-INF")
		metaInfFooExists = metaInfFooExists || strings.HasPrefix(header.Name, "META-INF-foo")

		i++
	}

	assert.True(t, exampleccExist, "src/example_cc/chaincode_example02.js does not exists in tar file")
	assert.True(t, eventMetaInfExists, "META-INF/event.json does not exists in tar file")
	assert.True(t, examplecc1MetaInfExists, "META-INF/example1.json does not exists in tar file")
	assert.False(t, fooMetaInfoExists, "invalid root directory found")
	assert.False(t, metaInfFooExists, "invalid root directory found")
}

// Test Package Go ChainCode
func TestEmptyCreate(t *testing.T) {

	_, err := NewCCPackage("")
	if err == nil {
		t.Fatal("Package Empty GoLang CC must return an error.")
	}
}

// Test Bad Package Path for ChainCode packaging
func TestBadPackagePathGoLangCC(t *testing.T) {
	_, err := NewCCPackage("github.com")
	if err == nil {
		t.Fatalf("error expected from Create %s", err)
	}
}

// Test isSource set to true for any go readable files used in ChainCode packaging
func TestIsSourcePath(t *testing.T) {
	keep = []string{}
	isSrcVal := isSource(filepath.Join(".."))

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
