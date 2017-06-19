/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package packager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
)

// Test Package Go ChainCode
func TestPackageGoLangCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../../../test/fixtures"))

	ccPackage, err := PackageGoLangCC("github.com")
	if err != nil {
		t.Fatalf("error from PackageGoLangCC %v", err)
	}

	r := bytes.NewReader(ccPackage)
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
		fmt.Println(header.Name)
		if header.Name == "src/github.com/example_cc/example_cc.go" {
			exampleccExist = true
		}
		i++
	}

	if !exampleccExist {
		t.Fatalf("src/github.com/example_cc/example_cc.go not exist in tar file")
	}

}
