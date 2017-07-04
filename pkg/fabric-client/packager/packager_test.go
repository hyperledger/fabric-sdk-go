package packager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
	"testing"
)

// Test Packager wrapper ChainCode packaging
func TestPackageCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../../../test/fixtures"))

	ccPackage, err := PackageCC("github.com", "")
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
func TestEmptyPackageCC(t *testing.T) {
	os.Setenv("GOPATH", "")

	_, err := PackageCC("", "")
	if err == nil {
		t.Fatalf("Package Empty GoLang CC must return an error.")
	}
}

// Test Package Go ChainCode
func TestUndefinedPackageCC(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd %v", err)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../../../test/fixtures"))

	_, err = PackageCC("github.com", "UndefinedCCType")
	if err == nil {
		t.Fatalf("Undefined package UndefinedCCType GoLang CC must return an error.")
	}
}
