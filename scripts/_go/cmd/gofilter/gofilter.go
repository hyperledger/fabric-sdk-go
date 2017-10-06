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
	"path"
	"strings"
)

func main() {
	filename := flag.String("filename", "", "Go source file to filter")
	filtersFlag := flag.String("filters", "", "filters to run [allowfn]")
	allowedFuncsFlag := flag.String("fn", "", "func names to allow - comma separated")
	flag.Parse()

	if len(*filename) == 0 || len(*filtersFlag) == 0 {
		fmt.Printf("Usage of %s:\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		return
	}

	allowedFuncs := strings.Split(*allowedFuncsFlag, ",")
	filterStrings := strings.Split(*filtersFlag, ",")

	// enabled filters
	filters := make(map[string]bool)
	for _, fs := range filterStrings {
		filters[fs] = true
	}

	// load filter table as an allow list
	ft := filterTbl{
		fn: make(map[string]bool),
	}
	for _, af := range allowedFuncs {
		ft.fn[af] = true
	}

	// load the AST
	fileset := token.NewFileSet()
	astFile, err := parser.ParseFile(fileset, *filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error paring file: %v", err)
		os.Exit(1)
	}

	// create an initial comment map (will be filtered later)
	cmap := ast.NewCommentMap(fileset, astFile, astFile.Comments)

	// filter the AST by func name
	tDecls := make([]ast.Decl, 0, len(astFile.Decls))
	for _, d := range astFile.Decls {
		if _, ok := d.(*ast.FuncDecl); ok {
			if filters["allowfn"] {
				tDecls = ft.applyFilterAllowFn(tDecls, d)
			}
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
		fmt.Fprintf(os.Stderr, "Failed to output filtered source code: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s", buf.Bytes())
}

type filterTbl struct {
	fn map[string]bool
}

func (t *filterTbl) applyFilterAllowFn(decls []ast.Decl, d ast.Decl) []ast.Decl {
	if ast.FilterDecl(d, t.filterFn) {
		decls = append(decls, d)
	}
	return decls
}

func (t *filterTbl) filterFn(name string) bool {
	return t.fn[name]
}
