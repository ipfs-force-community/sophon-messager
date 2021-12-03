package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
)

type metaInfo struct {
	node  *ast.Field
	ftype *ast.FuncType
}

type methodInfo struct {
	Name string
	Tag  string
}

type Visitor struct {
	Methods map[string]metaInfo
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	st, ok := node.(*ast.TypeSpec)
	if !ok {
		return v
	}

	iface, ok := st.Type.(*ast.InterfaceType)
	if !ok {
		return v
	}
	for _, m := range iface.Methods.List {
		switch ft := m.Type.(type) {
		/*	case *ast.Ident:
			v.Include[st.Name.Name] = append(v.Include[st.Name.Name], ft.Name)*/
		case *ast.FuncType:
			v.Methods[m.Names[0].Name] = metaInfo{
				node:  m,
				ftype: ft,
			}
		}
	}

	return v
}

func main() {
	fset := token.NewFileSet()
	src, err := parser.ParseFile(fset, "./api/client/api.go", nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	v := &Visitor{make(map[string]metaInfo)}
	ast.Walk(v, src)
	cmap := ast.NewCommentMap(fset, src, src.Comments)

	var methods []string //for code diff
	for name := range v.Methods {
		methods = append(methods, name)
	}

	var methodInfos []methodInfo
	for _, name := range methods {
		metaInfo := v.Methods[name]
		filteredComments := cmap.Filter(metaInfo.node).Comments()
		mInfo := methodInfo{
			Name: metaInfo.node.Names[0].Name,
		}
		if len(filteredComments) > 0 {
			tagstr := filteredComments[len(filteredComments)-1].List[0].Text
			tagstr = strings.TrimPrefix(tagstr, "//")
			tl := strings.Split(strings.TrimSpace(tagstr), " ")
			for _, ts := range tl {
				tf := strings.Split(ts, ":")
				if len(tf) != 2 {
					continue
				}
				if tf[0] != "perm" { // todo: allow more tag types
					continue
				}
				mInfo.Tag = tf[1]
			}
		}

		methodInfos = append(methodInfos, mInfo)
	}

	w := os.Stdout
	err = doTemplate(w, methodInfos, fileTemplate)
	if err != nil {
		fmt.Println(err)
	}
}

var fileTemplate = `package controller
var AuthMap = map[string]string {
{{range .}}"{{.Name}}":"{{.Tag}}",
{{end}}
}
`

func doTemplate(w io.Writer, info interface{}, templ string) error {
	t := template.Must(template.New("").
		Funcs(template.FuncMap{}).Parse(templ))

	return t.Execute(w, info)
}
