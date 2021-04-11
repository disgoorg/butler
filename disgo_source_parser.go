package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var Packages []*ast.Package

type SourceInfo struct {
	PackageName    string
	PackagePath    string
	IdentifierName string
	Comment        string
	Kind           string
}

var filenameRegex = regexp.MustCompile(`/[^/.]+\.go$`)

func findInPackages(packageName string, identifierName string) *SourceInfo {
	for _, pck := range Packages {
		if pck.Name == packageName {
			for filePath, file := range pck.Files {
				var lastPos token.Pos = 0
				for key, identifier := range file.Scope.Objects {
					if key == identifierName {
						comment := ""
						for _, cmt := range file.Comments {
							if cmt.Pos() >= lastPos && cmt.Pos() <= identifier.Pos() {
								comment += "\n" + cmt.Text()
							}
						}

						packagePath := strings.ReplaceAll(filePath, "\\", "/")
						packagePath = strings.Replace(packagePath, "../", "", 1)
						packagePath = filenameRegex.ReplaceAllString(packagePath, "")

						return &SourceInfo{
							PackageName:    pck.Name,
							PackagePath:    packagePath,
							IdentifierName: identifier.Name,
							Comment:        comment,
							Kind:           identifier.Kind.String(),
						}
					}
					lastPos = identifier.Pos()
				}
			}
		}
	}
	return nil
}

func loadPackages(path string) error {
	fs := token.NewFileSet()
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			packages, err := parser.ParseDir(fs, path, nil, parser.ParseComments|parser.AllErrors)
			if err != nil {
				return err
			}
			for _, pck := range packages {
				Packages = append(Packages, pck)
			}
		}
		return nil
	})
	return err
}
