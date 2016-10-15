package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"text/template"
)

var (
	srcpath = flag.String("path", "", "source code file to analyze")
	structs = flag.String("structs", "", "structs to analyze, should be comman separated")
	tag     = flag.String("tag", "db", "tag name to look")
	suffix  = flag.String("suf", "column", "suffix for gen file name")
)

func main() {
	flag.Parse()
	if *srcpath == "" || *structs == "" {
		log.Fatal("nothing todo exiting")

	}
	// get file name from path
	psl := strings.Split(*srcpath, "/")
	fname := psl[len(psl)-1]

	src, err := ioutil.ReadFile(*srcpath)
	if err != nil {
		log.Fatalf("read file err %+v\n", err)
	}

	gensrc, err := genColumnStruct(fname, *tag, string(src), strings.Split(*structs, ","))

	if err != nil {
		log.Fatalf("generation err %+v\n", err)
	}

	genSourcPath := strings.Join(psl[:len(psl)-1], "/") + "/" + strings.Replace(fname, ".go", "_"+*suffix+".go", 1)
	err = ioutil.WriteFile(genSourcPath, gensrc, 0644)
	if err != nil {
		log.Fatalf("write file err %+v\n", err)
	}

	log.Printf("new file generated %v\n", genSourcPath)

}

// Struct represents db column representation
// name is struct which it references
// fields contain information regarding column names
// according to tag provided
type StructColumn struct {
	Name   string
	Fields []Field
}

// Name is struct field name
// TagValue is value stored in defined tag
type Field struct {
	Name     string
	TagValue string
}

// genColumnStruct generates helper structs for defined source code structs and defined tag
// tag, file name, file source code and structs should be provided
// function returns fmt formated source code according to generated template
func genColumnStruct(fname, tagName string, input string, structs []string) ([]byte, error) {
	// key is struct name
	mapOfStructs := make(map[string]StructColumn)
	pkg, m, err := parse(fname, input, structs)
	if err != nil {
		return nil, err
	}

	for k, v := range m {
		st := StructColumn{}
		st.Name = k

		for _, fld := range v.Fields.List {
			tagVal, ok := tagLookup(fld.Tag.Value, tagName)
			if !ok || tagVal == "" {
				return nil, errors.New("value for defined tag: " + tagName + " not found")
			}
			st.Fields = append(st.Fields, Field{
				Name: fld.Names[0].Name,
				// get tag value we need
				TagValue: tagVal,
			})
		}
		mapOfStructs[st.Name] = st
	}
	// create data struct to pass it to template
	var data = struct {
		PackageName string
		Structs     map[string]StructColumn
	}{
		PackageName: pkg,
		Structs:     mapOfStructs,
	}

	var buf bytes.Buffer
	if err := generatedTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

// Parse source file and returns map[StructName]*ast.StructType for defined structs
func parse(fname, input string, structs []string) (string, map[string]*ast.StructType, error) {
	// package name
	pkg := ""
	m := make(map[string]*ast.StructType)
	// Create the AST by parsing src
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, fname, input, 0)
	if err != nil {
		fmt.Printf("parse err %+v\n", err) // output for debug

		return pkg, m, err
	}
	pkg = f.Name.Name
	// flags to check that we have "type Smth struct" declaration
	isTypeDecl := false
	typeName := ""
	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				// found type decl
				isTypeDecl = true
			}
		case *ast.Ident:
			if isTypeDecl {
				// found type name declared
				typeName = x.Name
			}

		case *ast.StructType:
			for i := range structs {
				// check if we have correct declaration
				// and struct which we need
				if typeName == structs[i] && isTypeDecl {
					m[typeName] = x
				}
			}
			// uncheck flags for other structs
			isTypeDecl = false
			typeName = ""
		}
		return true
	})
	return pkg, m, nil
}

// For robustness and consistence tag lookup func used from reflect
// std library https://golang.org/src/reflect/type.go
//
// Copyright 2009 The Go Authors. All rights reserved.
//
// Lookup returns the value associated with key in the tag string.
// If the key is present in the tag the value (which may be empty)
// is returned. Otherwise the returned value will be the empty string.
// The ok return value reports whether the value was explicitly set in
// the tag string. If the tag does not have the conventional format,
// the value returned by Lookup is unspecified.
func tagLookup(tag, key string) (value string, ok bool) {
	for tag != "" {
		// clean from raw quotes
		tag = strings.Trim(tag, "`")
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			value, err := strconv.Unquote(qvalue)
			if err != nil {
				break
			}
			return value, true
		}
	}
	return "", false
}

var generatedTmpl = template.Must(template.New("generated").Parse(columnTmpl))

var columnTmpl = `// generated by sqlhelp; DO NOT EDIT 
package {{.PackageName}}

{{range $typename, $struct := .Structs}}
type _{{$typename}}Column struct {
     {{range $fld := $struct.Fields}}
     {{$fld.Name}} string
     {{end}}
}
{{end}}

var (
{{range $typename, $struct := .Structs}}
    {{$typename}}Columns _{{$typename}}Column
{{end}}
)

func init() {
{{range $typename, $struct := .Structs}}
     // define {{$typename}} column names
     {{range $fld := $struct.Fields}}
     {{$typename}}Columns.{{$fld.Name}} = "{{$fld.TagValue}}"
     {{end}}
{{end}}
}
`
