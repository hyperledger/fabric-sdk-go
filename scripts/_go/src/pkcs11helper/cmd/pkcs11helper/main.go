/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	pw "pkcs11helper/pkg/pkcs11wrapper"

	"github.com/miekg/pkcs11"
)

var (
	defaultPkcs11LibPaths = []string{
		"/usr/lib/softhsm/libsofthsm2.so",
		"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
		"/usr/lib/s390x-linux-gnu/softhsm/libsofthsm2.so",
		"/usr/lib/powerpc64le-linux-gnu/softhsm/libsofthsm2.so",
		"/usr/local/Cellar/softhsm/2.1.0/lib/softhsm/libsofthsm2.so",
	}
)

func main() {

	// get flags
	pkcs11Library := flag.String("lib", "", "Location of pkcs11 library (Defaults to a list of possible paths to libsofthsm2.so)")
	slotLabel := flag.String("slot", "ForFabric", "Slot Label")
	slotPin := flag.String("pin", "98765432", "Slot PIN")
	action := flag.String("action", "list", "list,import")
	keyFile := flag.String("keyFile", "testdata/key.ec.pem", "path to pem encoded EC private key you want to import")

	flag.Parse()

	// initialize pkcs11
	var p11Lib string
	var err error

	if *pkcs11Library == "" {
		// if no lib is specified, just try to find libsofthsm2.so
		p11Lib, err = searchForLib(strings.Join(defaultPkcs11LibPaths, ","))
		exitWhenError(err)
	} else {
		p11Lib, err = searchForLib(*pkcs11Library)
		exitWhenError(err)
	}

	p11w := pw.Pkcs11Wrapper{
		Library: pw.Pkcs11Library{
			Path: p11Lib,
		},
		SlotLabel: *slotLabel,
		SlotPin:   *slotPin,
	}

	err = p11w.InitContext()
	exitWhenError(err)

	err = p11w.InitSession()
	exitWhenError(err)

	err = p11w.Login()
	exitWhenError(err)

	// defer cleanup
	defer p11w.CloseContext()

	// complete actions
	switch *action {

	case "import":
		err = p11w.ImportECKeyFromFile(*keyFile)
		exitWhenError(err)

	default:
		p11w.ListObjects(
			[]*pkcs11.Attribute{},
			50,
		)

	}

}

// exit properly
func exitWhenError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// return the first path that is found
func searchForLib(paths string) (firstFound string, err error) {

	libPaths := strings.Split(paths, ",")
	for _, path := range libPaths {
		if _, err = os.Stat(strings.TrimSpace(path)); !os.IsNotExist(err) {
			firstFound = strings.TrimSpace(path)
			break
		}
	}

	if firstFound == "" {
		err = fmt.Errorf("no suitable paths for pkcs11 library found: %s", paths)
	}

	return
}
