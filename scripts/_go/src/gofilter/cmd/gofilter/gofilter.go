/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ALLOW is the mode that allows a token in the program
const ALLOW = "allow"

func main() {
	filename := flag.String("filename", "", "Go source file to filter")
	filtersFlag := flag.String("filters", "", "filters to run [allowfn,allowgen,allowtype]")
	allowedFuncsFlag := flag.String("fn", "", "func names - comma separated")
	allowedGenFlag := flag.String("gen", "", "general decl names - comma separated")
	allowedTypeFlag := flag.String("type", "", "general decl types - comma separated")
	modeFlag := flag.String("mode", "allow", "allow or disallow")
	flag.Parse()

	if len(*filename) == 0 || len(*filtersFlag) == 0 {
		fmt.Printf("Usage of %s:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		return
	}

	// enabled filters
	filterStrings := strings.Split(*filtersFlag, ",")
	filters := make(map[string]bool)
	for _, fs := range filterStrings {
		filters[fs] = true
	}

	// load filter table as an allow list
	ft := filterTbl{
		mode: *modeFlag,
		fn:   make(map[string]bool),
		gen:  make(map[string]bool),
		typ:  make(map[int]bool),
	}

	if filters["fn"] && len(*allowedFuncsFlag) != 0 {
		allowedFuncs := strings.Split(*allowedFuncsFlag, ",")
		for _, af := range allowedFuncs {
			ft.fn[af] = true
		}
	}

	if filters["gen"] && len(*allowedGenFlag) != 0 {
		allowedGens := strings.Split(*allowedGenFlag, ",")
		for _, ag := range allowedGens {
			ft.gen[ag] = true
		}
	}

	if filters["type"] && len(*allowedTypeFlag) != 0 {
		allowedTypes := strings.Split(*allowedTypeFlag, ",")

		typeDict := make(map[string]int)
		typeDict["IMPORT"] = int(token.IMPORT)
		typeDict["CONST"] = int(token.CONST)

		for _, ats := range allowedTypes {
			if t, ok := typeDict[ats]; ok {
				ft.typ[t] = true
			} else {
				fmt.Fprintf(os.Stderr, "type is unknown: %s\n", ats)
				os.Exit(1)
			}
		}
	}

	// load the AST
	fileset := token.NewFileSet()
	astFile, err := parser.ParseFile(fileset, *filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %s\n", err)
		os.Exit(1)
	}

	// create an initial comment map (will be filtered later)
	cmap := ast.NewCommentMap(fileset, astFile, astFile.Comments)

	// filter the AST by func name
	tDecls := make([]ast.Decl, 0, len(astFile.Decls))
	for _, d := range astFile.Decls {
		if _, ok := d.(*ast.FuncDecl); ok && filters["fn"] {
			tDecls = ft.applyFilterAllowFn(tDecls, d)
		} else if g, ok := d.(*ast.GenDecl); ok && filters["type"] && ft.typ[int(g.Tok)] {
			if ft.mode == ALLOW {
				tDecls = append(tDecls, d)
			}
		} else if _, ok := d.(*ast.GenDecl); ok && filters["gen"] {
			tDecls = ft.applyFilterAllowGen(tDecls, d)
		} else {
			tDecls = append(tDecls, d)
		}
	}
	astFile.Decls = tDecls

	// filter out comments attached to removed functions
	astFile.Comments = cmap.Filter(astFile).Comments()

	// output the filtered source code
	var buf bytes.Buffer
	if err := format.Node(&buf, fileset, astFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to output filtered source code: %s", err)
		os.Exit(1)
	}
	fmt.Printf("%s", buf.Bytes())
}

type filterTbl struct {
	mode string
	fn   map[string]bool
	gen  map[string]bool
	typ  map[int]bool
}

func (t *filterTbl) applyFilterAllowFn(decls []ast.Decl, d ast.Decl) []ast.Decl {
	if ast.FilterDecl(d, t.filterFn) {
		decls = append(decls, d)
	}
	return decls
}

func (t *filterTbl) applyFilterAllowGen(decls []ast.Decl, d ast.Decl) []ast.Decl {
	if ast.FilterDecl(d, t.filterGen) {
		decls = append(decls, d)
	}
	return decls
}

func (t *filterTbl) filterFn(name string) bool {
	return t.fn[name] == (t.mode == ALLOW)
}

func (t *filterTbl) filterGen(name string) bool {
	return t.gen[name] == (t.mode == ALLOW)
}
