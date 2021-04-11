package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var Packages []*ast.Package

type SourceInfo struct {
	PackageName    string
	PackagePath    string
	IdentifierName string
	Comment        string
	Kind           ast.ObjKind
	Params         *string
	Results        *string
}

var filenameRegex = regexp.MustCompile(`/[^/.]+\.go$`)

func findInPackages(packageName string, identifierName string) *SourceInfo {
	for _, pck := range Packages {
		if pck.Name == packageName {
			for filePath, file := range pck.Files {
				var lastPos token.Pos = 0
				// make slice from file.Scope.Objects & sort after Pos
				objects := make([]*ast.Object, len(file.Scope.Objects))
				i := 0
				for _, obj := range file.Scope.Objects {
					objects[i] = obj
					i++
				}
				sort.SliceStable(objects, func(i, j int) bool {
					return objects[i].Pos() < objects[j].Pos()
				})

				for _, obj := range objects {
					if obj.Name == identifierName {
						comment := ""
						for _, cmt := range file.Comments {
							if cmt.Pos() > lastPos && cmt.Pos() < obj.Pos() {
								comment += "\n" + cmt.Text()
							}
						}

						packagePath := strings.ReplaceAll(filePath, "\\", "/")
						packagePath = strings.Replace(packagePath, "../", "", 1)
						packagePath = filenameRegex.ReplaceAllString(packagePath, "")

						sourceInfo := &SourceInfo{
							PackageName:    pck.Name,
							PackagePath:    packagePath,
							IdentifierName: obj.Name,
							Comment:        comment,
							Kind:           obj.Kind,
						}
						if obj.Kind == ast.Fun {
							fDecl := obj.Decl.(*ast.FuncDecl)

							params := ""
							for i, field := range fDecl.Type.Params.List {
								if i != 0 {
									params += ", "
								}
								if len(field.Names) > 0 {
									params += field.Names[0].Name + " "
								}
								expr := field.Type.(*ast.SelectorExpr)
								params += expr.X.(*ast.Ident).Name + "." + expr.Sel.Name
							}
							if params != "" {
								sourceInfo.Params = &params
							}

							results := ""
							for i, field := range fDecl.Type.Results.List {
								if i != 0 {
									results += ", "
								}
								if len(field.Names) > 0 {
									results += field.Names[0].Name + " "
								}
								switch x := field.Type.(type) {
								case *ast.SelectorExpr:
									results += x.X.(*ast.Ident).Name + "." + x.Sel.Name
								case *ast.Ident:
									results += x.Name
								}

							}
							if results != "" {
								sourceInfo.Results = &results
							}
						}
						return sourceInfo
					}
					lastPos = obj.Pos()
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
