package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func ParseFile(filePath string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, filePath, nil, parser.ParseComments)
}

// parseFuncCmt 解析方法上的注释
func parseFuncCmt(funcDecl *ast.FuncDecl, title string) (ans []string) {
	if funcDecl.Doc == nil {
		return
	}
	comment := funcDecl.Doc.Text()
	if !strings.Contains(comment, title) {
		return
	}
	for _, cmt := range funcDecl.Doc.List {
		cmtLine := cmt.Text
		if !strings.Contains(cmtLine, title) {
			continue
		}
		ans = append(ans, cmtLine)
	}
	return ans
}
