// Copyright (c) 2012-2018 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

// codecgen generates static implementations of the encoder and decoder functions
// for a given type, bypassing reflection, and giving some performance benefits in terms of
// wall and cpu time, and memory usage.
//
// Benchmarks (as of Dec 2018) show that codecgen gives about
//
//   - for binary formats (cbor, etc): 25% on encoding and 30% on decoding to/from []byte
//   - for text formats (json, etc): 15% on encoding and 25% on decoding to/from []byte
//
// # Note that (as of Dec 2018) codecgen completely ignores
//
//   - MissingFielder interface
//     (if you types implements it, codecgen ignores that)
//   - decode option PreferArrayOverSlice
//     (we cannot dynamically create non-static arrays without reflection)
//
// In explicit package terms: codecgen generates codec.Selfer implementations for a set of types.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const genCodecPkg = "codec1978" // keep this in sync with codec.genCodecPkg

const genFrunMainTmpl = `//+build ignore

// Code generated - temporary main package for codecgen - DO NOT EDIT.

package main
{{ if .Types }}import "{{ .ImportPath }}"{{ end }}
func main() {
	{{ $.PackageName }}.CodecGenTempWrite{{ .RandString }}()
}
`

// const genFrunPkgTmpl = `//+build codecgen
const genFrunPkgTmpl = `

// Code generated - temporary package for codecgen - DO NOT EDIT.

package {{ $.PackageName }}

import (
	{{ if not .CodecPkgFiles }}{{ .CodecPkgName }} "{{ .CodecImportPath }}"{{ end }}
	"os"
	"reflect"
	"bytes"
	"strings"
	"go/format"
)

func CodecGenTempWrite{{ .RandString }}() {
	os.Remove("{{ .OutFile }}")
	fout, err := os.Create("{{ .OutFile }}")
	if err != nil {
		panic(err)
	}
	defer fout.Close()
	
	var typs []reflect.Type
	var typ reflect.Type
	var numfields int
{{ range $index, $element := .Types }}
	var t{{ $index }} {{ . }}
typ = reflect.TypeOf(t{{ $index }})
	typs = append(typs, typ)
	if typ.Kind() == reflect.Struct { numfields += typ.NumField() } else { numfields += 1 }
{{ end }}

	// println("initializing {{ .OutFile }}, buf size: {{ .AllFilesSize }}*16",
	// 	{{ .AllFilesSize }}*16, "num fields: ", numfields)
	var out = bytes.NewBuffer(make([]byte, 0, numfields*1024)) // {{ .AllFilesSize }}*16
	{{ if not .CodecPkgFiles }}{{ .CodecPkgName }}.{{ end }}Gen(out,
		"{{ .BuildTag }}", "{{ .PackageName }}", "{{ .RandString }}", {{ .NoExtensions }},
		{{ if not .CodecPkgFiles }}{{ .CodecPkgName }}.{{ end }}NewTypeInfos(strings.Split("{{ .StructTags }}", ",")),
		 typs...)

	bout, err := format.Source(out.Bytes())
	// println("... lengths: before formatting: ", len(out.Bytes()), ", after formatting", len(bout))
	if err != nil {
		fout.Write(out.Bytes())
		panic(err)
	}
	fout.Write(bout)
}

`

// Generate is given a list of *.go files to parse, and an output file (fout).
//
// It finds all types T in the files, and it creates 2 tmp files (frun).
//   - main package file passed to 'go run'
//   - package level file which calls *genRunner.Selfer to write Selfer impls for each T.
//
// We use a package level file so that it can reference unexported types in the package being worked on.
// Tool then executes: "go run __frun__" which creates fout.
// fout contains Codec(En|De)codeSelf implementations for every type T.
func Generate(outfile, buildTag, codecPkgPath string,
	uid int64,
	goRunTag string, st string,
	regexName, notRegexName *regexp.Regexp,
	deleteTempFile, noExtensions bool,
	infiles ...string) (err error) {
	// For each file, grab AST, find each type, and write a call to it.
	if len(infiles) == 0 {
		return
	}
	if codecPkgPath == "" {
		return errors.New("codec package path cannot be blank")
	}
	if outfile == "" {
		return errors.New("outfile cannot be blank")
	}
	if uid < 0 {
		uid = -uid
	} else if uid == 0 {
		rr := rand.New(rand.NewSource(time.Now().UnixNano()))
		uid = 101 + rr.Int63n(9777)
	}
	// We have to parse dir for package, before opening the temp file for writing (else ImportDir fails).
	// Also, ImportDir(...) must take an absolute path.
	lastdir := filepath.Dir(outfile)
	absdir, err := filepath.Abs(lastdir)
	if err != nil {
		return
	}
	importPath, err := pkgPath(absdir)
	if err != nil {
		return
	}
	type tmplT struct {
		CodecPkgName    string
		CodecImportPath string
		ImportPath      string
		OutFile         string
		PackageName     string
		RandString      string
		BuildTag        string
		StructTags      string
		Types           []string
		AllFilesSize    int64
		CodecPkgFiles   bool
		NoExtensions    bool
	}
	tv := tmplT{
		CodecPkgName:    genCodecPkg,
		OutFile:         outfile,
		CodecImportPath: codecPkgPath,
		BuildTag:        buildTag,
		RandString:      strconv.FormatInt(uid, 10),
		StructTags:      st,
		NoExtensions:    noExtensions,
	}
	tv.ImportPath = importPath
	if tv.ImportPath == tv.CodecImportPath {
		tv.CodecPkgFiles = true
		tv.CodecPkgName = "codec"
	} else {
		// HACK: always handle vendoring. It should be typically on in go 1.6, 1.7
		tv.ImportPath = stripVendor(tv.ImportPath)
	}
	astfiles := make([]*ast.File, len(infiles))
	var fi os.FileInfo
	for i, infile := range infiles {
		if filepath.Dir(infile) != lastdir {
			err = errors.New("all input files must all be in same directory as output file")
			return
		}
		if fi, err = os.Stat(infile); err != nil {
			return
		}
		tv.AllFilesSize += fi.Size()

		fset := token.NewFileSet()
		astfiles[i], err = parser.ParseFile(fset, infile, nil, 0)
		if err != nil {
			return
		}
		if i == 0 {
			tv.PackageName = astfiles[i].Name.Name
			if tv.PackageName == "main" {
				// codecgen cannot be run on types in the 'main' package.
				// A temporary 'main' package must be created, and should reference the fully built
				// package containing the types.
				// Also, the temporary main package will conflict with the main package which already has a main method.
				err = errors.New("codecgen cannot be run on types in the 'main' package")
				return
			}
		}
	}

	// keep track of types with selfer methods
	// selferMethods := []string{"CodecEncodeSelf", "CodecDecodeSelf"}
	selferEncTyps := make(map[string]bool)
	selferDecTyps := make(map[string]bool)
	for _, f := range astfiles {
		for _, d := range f.Decls {
			// if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv != nil && fd.Recv.NumFields() == 1 {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv != nil && len(fd.Recv.List) == 1 {
				recvType := fd.Recv.List[0].Type
				if ptr, ok := recvType.(*ast.StarExpr); ok {
					recvType = ptr.X
				}
				if id, ok := recvType.(*ast.Ident); ok {
					switch fd.Name.Name {
					case "CodecEncodeSelf":
						selferEncTyps[id.Name] = true
					case "CodecDecodeSelf":
						selferDecTyps[id.Name] = true
					}
				}
			}
		}
	}

	// now find types
	for _, f := range astfiles {
		for _, d := range f.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				for _, dd := range gd.Specs {
					if td, ok := dd.(*ast.TypeSpec); ok {
						// if len(td.Name.Name) == 0 || td.Name.Name[0] > 'Z' || td.Name.Name[0] < 'A' {
						if len(td.Name.Name) == 0 {
							continue
						}

						// only generate for:
						//   struct: StructType
						//   primitives (numbers, bool, string): Ident
						//   map: MapType
						//   slice, array: ArrayType
						//   chan: ChanType
						// do not generate:
						//   FuncType, InterfaceType, StarExpr (ptr), etc
						//
						// We generate for all these types (not just structs), because they may be a field
						// in another struct which doesn't have codecgen run on it, and it will be nice
						// to take advantage of the fact that the type is a Selfer.
						switch td.Type.(type) {
						case *ast.StructType, *ast.Ident, *ast.MapType, *ast.ArrayType, *ast.ChanType:
							// only add to tv.Types iff
							//   - it matches per the -r parameter
							//   - it doesn't match per the -nr parameter
							//   - it doesn't have any of the Selfer methods in the file
							if regexName.FindStringIndex(td.Name.Name) != nil &&
								notRegexName.FindStringIndex(td.Name.Name) == nil &&
								!selferEncTyps[td.Name.Name] &&
								!selferDecTyps[td.Name.Name] {
								tv.Types = append(tv.Types, td.Name.Name)
							}
						}
					}
				}
			}
		}
	}

	if len(tv.Types) == 0 {
		return
	}

	// we cannot use ioutil.TempFile, because we cannot guarantee the file suffix (.go).
	// Also, we cannot create file in temp directory,
	// because go run will not work (as it needs to see the types here).
	// Consequently, create the temp file in the current directory, and remove when done.

	// frun, err = ioutil.TempFile("", "codecgen-")
	// frunName := filepath.Join(os.TempDir(), "codecgen-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".go")

	frunMainName := filepath.Join(lastdir, "codecgen-main-"+tv.RandString+".generated.go")
	frunPkgName := filepath.Join(lastdir, "codecgen-pkg-"+tv.RandString+".generated.go")
	if deleteTempFile {
		defer os.Remove(frunMainName)
		defer os.Remove(frunPkgName)
	}
	// var frunMain, frunPkg *os.File
	if _, err = gen1(frunMainName, genFrunMainTmpl, &tv); err != nil {
		return
	}
	if _, err = gen1(frunPkgName, genFrunPkgTmpl, &tv); err != nil {
		return
	}

	// remove outfile, so "go run ..." will not think that types in outfile already exist.
	os.Remove(outfile)

	// execute go run frun
	cmd := exec.Command("go", "run", "-tags", "codecgen.exec safe "+goRunTag, frunMainName) //, frunPkg.Name())
	cmd.Dir = lastdir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("error running 'go run %s': %v, console: %s",
			frunMainName, err, buf.Bytes())
		return
	}
	os.Stdout.Write(buf.Bytes())
	return
}

func gen1(frunName, tmplStr string, tv interface{}) (frun *os.File, err error) {
	os.Remove(frunName)
	if frun, err = os.Create(frunName); err != nil {
		return
	}
	defer frun.Close()

	t := template.New("")
	if t, err = t.Parse(tmplStr); err != nil {
		return
	}
	bw := bufio.NewWriter(frun)
	if err = t.Execute(bw, tv); err != nil {
		bw.Flush()
		return
	}
	if err = bw.Flush(); err != nil {
		return
	}
	return
}

// copied from ../gen.go (keep in sync).
func stripVendor(s string) string {
	// HACK: Misbehaviour occurs in go 1.5. May have to re-visit this later.
	// if s contains /vendor/ OR startsWith vendor/, then return everything after it.
	const vendorStart = "vendor/"
	const vendorInline = "/vendor/"
	if i := strings.LastIndex(s, vendorInline); i >= 0 {
		s = s[i+len(vendorInline):]
	} else if strings.HasPrefix(s, vendorStart) {
		s = s[len(vendorStart):]
	}
	return s
}

func main() {
	o := flag.String("o", "", "out file")
	c := flag.String("c", genCodecPath, "codec path")
	t := flag.String("t", "", "build tag to put in file")
	r := flag.String("r", ".*", "regex for type name to match")
	nr := flag.String("nr", "^$", "regex for type name to exclude")
	rt := flag.String("rt", "", "tags for go run")
	st := flag.String("st", "codec,json", "struct tag keys to introspect")
	x := flag.Bool("x", false, "keep temp file")
	_ = flag.Bool("u", false, "Allow unsafe use. ***IGNORED*** - kept for backwards compatibility: ")
	d := flag.Int64("d", 0, "random identifier for use in generated code")
	nx := flag.Bool("nx", false, "do not support extensions - support of extensions may cause extra allocation")

	flag.Parse()
	err := Generate(*o, *t, *c, *d, *rt, *st,
		regexp.MustCompile(*r), regexp.MustCompile(*nr), !*x, *nx, flag.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codecgen error: %v\n", err)
		os.Exit(1)
	}
}
