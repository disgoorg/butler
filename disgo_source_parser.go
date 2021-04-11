package main

import (
	"archive/zip"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/DisgoOrg/disgo/api"
)

const disgoURL = "https://api.github.com/repos/DisgoOrg/disgo/releases/latest"

var disgoGitInfo *GitInfo

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

func loadPackages() error {
	fs := token.NewFileSet()
	err := filepath.Walk("disgo", func(path string, info os.FileInfo, err error) error {
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
	//_ = os.Remove("disgo.zip")
	//_ = os.RemoveAll("disgo")
	return err
}

func downloadDisgo(restClient api.RestClient) error {
	gitInfo, err := getGitInfo(restClient)
	disgoGitInfo = gitInfo

	rs, err := http.Get(disgoGitInfo.ZipballUrl)
	if err != nil {
		return err
	}

	defer func() {
		_ = rs.Body.Close()
	}()

	if rs.StatusCode != http.StatusOK {
		return errors.New("no status code 200")
	}

	out, err := os.Create("disgo.zip")
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, rs.Body)
	if err != nil {
		return err
	}

	return unzip("./disgo.zip")
}

// https://stackoverflow.com/a/65618964
func unzip(source string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	dirName := reader.File[0].Name
	for _, file := range reader.File {
		if file.Mode().IsDir() {
			continue
		}
		fileName := strings.Replace(file.Name, dirName, "disgo", 1)
		err = os.MkdirAll(path.Dir(fileName), os.ModeDir)
		if err != nil {
			return err
		}
		open, err := file.Open()
		if err != nil {
			return err
		}
		create, err := os.Create(fileName)
		if err != nil {
			return err
		}
		_, err = create.ReadFrom(open)
		if err != nil {
			return err
		}
	}
	return nil
}
