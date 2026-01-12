package main

import (
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"go/ast"
	"log"
	"os"
	"strings"
	"text/template"
)

const (
	PermissionTag = "@Permission"
	RouterTag     = "@Router"
)

//go:embed  relation.gtpl
var ObjectPoolTemplateText string

var permissionResourceTemplate = template.Must(template.New("relation.gtpl").Parse(ObjectPoolTemplateText))

type ProjectParser struct {
	ServiceName string
	BaseRoute   string
	SourcePath  string
	Dest        string
	DestPackage string
}

type PermissionAnnotation struct {
	RouterCmt     string   `json:"router_cmt"`
	PermissionCmt []string `json:"permission_cmt"`
}

func (p *PermissionAnnotation) routerTags() []string {
	rs := strings.SplitAfter(p.RouterCmt, RouterTag)
	rts := strings.Split(rs[1], " ")
	rts = lo.Filter(rts, func(item string, index int) bool { return item != "" })
	return rts
}

func (p *PermissionAnnotation) permissionTags() (results [][]string) {
	for i := range p.PermissionCmt {
		ps := strings.SplitAfter(p.PermissionCmt[i], PermissionTag)
		pts := strings.Split(ps[1], " ")
		pts = lo.Filter(pts, func(item string, index int) bool { return item != "" })

		results = append(results, pts)
	}
	return results
}

type PermissionResource struct {
	ServiceName string `json:"service_name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
}

type PermissionResourceSlice struct {
	Annos       []*PermissionResource
	PackageName string
	ServiceName string `json:"service_name"`
}

func NewProjectParser(path, serviceName, baseRoute, dest, destPackage string) *ProjectParser {
	return &ProjectParser{
		SourcePath:  path,
		ServiceName: serviceName,
		BaseRoute:   baseRoute,
		DestPackage: destPackage,
		Dest:        dest,
	}
}

func (p *ProjectParser) readAllFile() ([]string, error) {
	files := make([]string, 0)
	if PathExists(p.SourcePath) {
		allFiles, err := GetAllGoFiles(p.SourcePath)
		if err != nil {
			return nil, err
		}
		files = allFiles
	} else {
		files = append(files, p.SourcePath)
	}
	return files, nil
}

func (p *ProjectParser) ReadAllReadAnnotation() (*PermissionResourceSlice, error) {
	files, err := p.readAllFile()
	if err != nil {
		return nil, err
	}
	annos := make([]*PermissionResource, 0)
	for _, file := range files {
		ans, err := p.ReadAnnotationSlice(file)
		if err != nil {
			return nil, fmt.Errorf("读取%s,注解错误:%s", file, err.Error())
		}
		if len(ans) > 0 {
			annos = append(annos, ans...)
		}
	}
	return &PermissionResourceSlice{
		PackageName: p.DestPackage,
		ServiceName: p.ServiceName,
		Annos:       annos,
	}, nil
}

func (p *ProjectParser) ReadAnnotationSlice(filePath string) ([]*PermissionResource, error) {
	astFile, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	annos := make([]*PermissionResource, 0)
	for _, decl := range astFile.Decls {
		switch f := decl.(type) {
		case *ast.FuncDecl:
			routerCmts := parseFuncCmt(f, RouterTag)
			permissionCmts := parseFuncCmt(f, PermissionTag)
			if len(routerCmts) > 0 && len(permissionCmts) > 0 {
				pa := &PermissionAnnotation{
					RouterCmt:     routerCmts[0],
					PermissionCmt: permissionCmts,
				}
				annos = append(annos, pa.parseCmt(p.ServiceName, p.BaseRoute)...)
			}
		}
	}
	return annos, nil
}

func (p *PermissionAnnotation) parseCmt(serviceName, baseRoute string) (results []*PermissionResource) {
	//routerTag
	rts := p.routerTags()
	ptss := p.permissionTags()

	for i := range ptss {
		pts := ptss[i]
		results = append(results, &PermissionResource{
			ServiceName: serviceName,
			Path:        baseRoute + replaceMatcher(strings.TrimSpace(rts[0])),
			Method:      strings.ToUpper(strings.Trim(rts[1], "[]")),
			Resource:    pts[0],
			Action:      pts[1],
		})
	}
	return results
}

func WriteTemplate(tmpl *template.Template, app any, filePath string) error {
	destFile, err := os.Create(filePath)
	if err != nil {
		log.Println("Open Files Error:", err)
		return err
	}
	return tmpl.Execute(destFile, app)
}

// replaceMatcher 将匹配{}换:， 当前只支持这种简单的匹配
func replaceMatcher(p string) string {
	p = strings.ReplaceAll(p, "}", "")
	p = strings.ReplaceAll(p, "{", ":")
	return p
}
