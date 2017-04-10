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
	os.Setenv("GOPATH", path.Join(pwd, "../../test/fixtures"))

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
