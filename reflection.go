package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type StructStore map[string]*Struct
type ImportStore map[string]*Import

type Import struct {
	ImplicitName string
	ExplicitName string
	Path string
}

type File struct {
	Path string
	Imports ImportStore
	Package *packages.Package
}

type Struct struct {
	File *File
	Name string
	FullName string
	PublicMethods []Method
}

type Method struct {
	Name string
	Receiver *Param
	Params []Param
	ReturnType []Param
}

type Param struct {
	Name string
	IsPtr bool
	TypeName string
	FullTypeName string
}


type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}

func main() {
	structs := Parse("reftest")
	GenWrappers(structs)
}

func GenWrappers(structs StructStore) {
	wrappedStructsByFile := map[string][]*Struct{}
	for _, st := range structs {
		if len(st.PublicMethods) > 0 {
			if _, ok := wrappedStructsByFile[st.File.Path]; !ok {
				wrappedStructsByFile[st.File.Path] = []*Struct{}
			}
			wrappedStructsByFile[st.File.Path] = append(wrappedStructsByFile[st.File.Path], st)
		}
	}

	for filePath, wrappedStructs := range wrappedStructsByFile {
		output := bytes.NewBufferString("")
		printWrappers(output, wrappedStructs[0].File, wrappedStructs, structs)
		fmt.Println("file:", filePath)
		fmt.Println(output.String())
		fmt.Println()
		fmt.Println()
	}
}

func typeWithImportName(fullTypeName string) string {
	parts := strings.Split(fullTypeName, "/")
	return parts[len(parts) - 1]
}

func isWrappedStruct(st *Struct) bool {
	return len(st.PublicMethods) > 0
}

func formatParams(params []Param, f *File, structs StructStore) string {
	var strs []string

	for _, p := range params {
		typeName := p.TypeName
		if st, ok := structs[p.FullTypeName]; ok {
			if !isWrappedStruct(st) {
				typeName = "orig_" + typeWithImportName(p.FullTypeName)
			} else if st.File.Path != f.Path {
				typeName = typeWithImportName(p.FullTypeName)
			}
		}
		if p.IsPtr {
			typeName = "*" + typeName
		}

		if p.Name != "" {
			strs = append(strs, fmt.Sprintf("%s %s", p.Name, typeName))
		} else {
			strs = append(strs, fmt.Sprintf("%s", typeName))
		}
	}
	return strings.Join(strs, ", ")
}

func formatParamsCall(params []Param) string {
	var strs []string
	for _, p := range params {
		strs = append(strs, p.Name)
	}
	return strings.Join(strs, ", ")
}

func formatReceiver(p *Param) string {
	if p == nil {
		panic(fmt.Errorf("must have receiver"))
	}
	// Always use a ptr receiver
	return fmt.Sprintf("p *%s", p.TypeName)
}

func printWrappers(output io.Writer, file *File, wrappedStructs []*Struct, structs StructStore) {
	for _, imp := range file.Imports {
		if imp.ExplicitName != "" {
			_, _ = fmt.Fprintf(output, "import orig_%s \"%s\"\n", imp.ExplicitName, imp.Path)
		} else {
			_, _ = fmt.Fprintf(output, "import orig_%s \"%s\"\n", imp.ImplicitName, imp.Path)
		}
	}
	_, _ = fmt.Fprintf(output, "import orig_%s \"%s\"\n\n", file.Package.Name, file.Package.PkgPath)

	for _, st := range wrappedStructs {
		printWrapper(output, st, structs)
	}
}

func printWrapper(output io.Writer, st *Struct, structs StructStore) {
	nameParts := strings.Split(st.Name, ".")
	interfaceName := nameParts[len(nameParts) - 1]

	_, _ = fmt.Fprintf(output, "type %s interface {\n", interfaceName)
	for _, m := range st.PublicMethods {
		_, _ = fmt.Fprintf(output, "\t%s(%s) (%s)\n", m.Name, formatParams(m.Params, st.File, structs), formatParams(m.ReturnType, st.File, structs))
	}
	_, _ = fmt.Fprintf(output, "}\n\n")

	structName := fmt.Sprintf("%s%sWrapper", strings.ToLower(st.Name[0:1]), st.Name[1:])

	implTypeName := typeWithImportName(st.FullName)
	_, _ = fmt.Fprintf(output, "type %s struct {\n", structName)
	_, _ = fmt.Fprintf(output, "\timpl %s\n", implTypeName)
	_, _ = fmt.Fprintf(output, "}\n\n")

	_, _ = fmt.Fprintf(output, "func New%s(impl %s) %s {\n", interfaceName, implTypeName, interfaceName)
	_, _ = fmt.Fprintf(output, "\treturn &%s{impl: impl}\n", structName)
	_, _ = fmt.Fprintf(output, "}\n\n")

	for _, m := range st.PublicMethods {
		if len(m.ReturnType) > 0 {
			_, _ = fmt.Fprintf(output, "func (%s) %s(%s) (%s) {\n",
				formatReceiver(m.Receiver),
				m.Name,
				formatParams(m.Params, st.File, structs),
				formatParams(m.ReturnType, st.File, structs))
			_, _ = fmt.Fprintf(output, "\treturn impl.%s(%s)\n", m.Name, formatParamsCall(m.Params))
			_, _ = fmt.Fprintf(output, "}\n\n")
		} else {
			_, _ = fmt.Fprintf(output, "func (%s) %s(%s) {\n",
				formatReceiver(m.Receiver),
				m.Name,
				formatParams(m.Params, st.File, structs))
			_, _ = fmt.Fprintf(output, "\timpl.%s(%s)\n", m.Name, formatParamsCall(m.Params))
			_, _ = fmt.Fprintf(output, "}\n\n")
		}
	}
}

//func print(structs StructStore) {
//	for _, st := range structs {
//		fmt.Println("struct:", st.FullName)
//		for _, m := range st.PublicMethods {
//			fmt.Print("\tfunc (")
//			if m.Receiver != nil {
//				fmt.Print(m.Receiver.Name, m.Receiver.FullTypeName)
//			}
//			fmt.Printf(") %s(", m.Name)
//			for _, p := range m.Params {
//				fmt.Printf("%s %s, ", p.Name, p.FullTypeName)
//			}
//			fmt.Print(") (")
//			for _, p := range m.ReturnType {
//				fmt.Printf("%s, ", p.FullTypeName)
//			}
//			fmt.Println(") {}")
//		}
//	}
//}

func Parse(dirs... string) StructStore {
	structs := StructStore{}

	// Structs
	fset := token.NewFileSet()
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
				parseStructsFromFile(fset, path, filepath.Dir(path), structs)
			}
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	// Methods
	fset = token.NewFileSet()
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
				parseMethodsFromFile(fset, path, filepath.Dir(path), structs)
			}
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	return structs
}

func createImportStore(f *ast.File) ImportStore {
	imports := ImportStore{}
	for _, astImp := range f.Imports {
		name, imp := newImport(astImp)
		imports[name] = imp
	}
	return imports
}

func parseMethodsFromFile(fset *token.FileSet, filepath string, directory string, structs StructStore) {
	fmt.Println("Parsing file for methods:", filepath)
	src, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	fillMethods(f, directory, structs, createImportStore(f))
}

func newImport(imp *ast.ImportSpec) (string, *Import) {
	path := imp.Path.Value[1:len(imp.Path.Value) - 1]
	parts := strings.Split(path, "/")

	newImport := &Import{Path: path, ImplicitName: parts[len(parts) - 1]}
	if imp.Name != nil {
		newImport.ExplicitName = imp.Name.Name
		return newImport.ExplicitName, newImport
	}

	return newImport.ImplicitName, newImport
}

func parseStructsFromFile(fset *token.FileSet, filepath string, directory string, structs StructStore) {
	fmt.Println("Parsing file for structs:", filepath)
	src, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	fillStructs(f, directory, structs, &File{Path:filepath, Imports: createImportStore(f), Package:getPkg(directory)})
}

func fillMethods(node ast.Node, directory string, structs StructStore, store ImportStore) {
	pkg := getPkg(directory)

	ast.Walk(walker(func(node ast.Node) bool {
		switch v := node.(type) {
		case *ast.FuncDecl:
			m := newMethod(v, structs, pkg, store)
			if m.Receiver != nil && strings.ToUpper(m.Name[0:1]) == m.Name[0:1] {
				s := structs[m.Receiver.FullTypeName]
				s.PublicMethods = append(s.PublicMethods, m)
			}
		}
		return true
	}), node)
}

func getPkg(directory string) *packages.Package {
	fullPath, err := filepath.Abs(directory)
	if err != nil {
		panic(err)
	}
	pkgs, err := packages.Load(nil, fullPath)
	if err != nil {
		panic(err)
	} else if (len(pkgs) != 1) {
		panic(fmt.Sprintf("found multiple packages for directory: %s", directory))
	}
	return pkgs[0]
}

func fillStructs(node ast.Node, directory string, structs StructStore, file *File) {
	ast.Walk(walker(func(node ast.Node) bool {
		switch v := node.(type) {
		case *ast.TypeSpec:
			if _, ok := v.Type.(*ast.StructType); ok {
				s := newStruct(v, file)
				structs[s.FullName] = s
			}
		}
		return true
	}), node)
}

func newStruct(v *ast.TypeSpec, file *File) *Struct {
	return &Struct{
		File: file,
		Name: v.Name.Name,
		FullName: fullTypeName(file.Package, v.Name.Name),
		PublicMethods: []Method{},
	}
}

func newMethod(v *ast.FuncDecl, structs StructStore, pkg *packages.Package, imports ImportStore) Method {
	return Method{
		Name: v.Name.Name,
		Receiver: maybeNewReceiver(v, structs, pkg),
		Params: getParams(v, structs, pkg, imports),
		ReturnType: getReturnParams(v, structs, pkg, imports),
	}
}

func getReturnParams(v *ast.FuncDecl, structs StructStore, pkg *packages.Package, imports ImportStore) []Param {
	var params []Param
	for _, p := range v.Type.Results.List {
		fullTypeName, typeName, isPtr := getTypeName(p.Type, structs, pkg, imports)
		if p.Names != nil {
			for _, pn := range p.Names {
				params = append(params, Param{Name: pn.Name, IsPtr: isPtr, TypeName: typeName, FullTypeName: fullTypeName})
			}
		} else {
			params = append(params, Param{IsPtr:isPtr, TypeName: typeName, FullTypeName:typeName})
		}
	}
	return params
}

func getParams(v *ast.FuncDecl, structs StructStore, pkg *packages.Package, imports ImportStore) []Param {
	var params []Param
	for _, p := range v.Type.Params.List {
		fullTypeName, typeName, isPtr := getTypeName(p.Type, structs, pkg, imports)
		for _, pn := range p.Names {
			params = append(params, Param{Name: pn.Name, IsPtr: isPtr, TypeName: typeName, FullTypeName: fullTypeName})
		}
	}
	return params
}

func getTypeName(exp ast.Expr, structs StructStore, pkg *packages.Package, imports ImportStore) (string, string, bool) {
	var fullName string
	var name string
	var isPtr bool
	switch xv := exp.(type) {
	case *ast.StarExpr:
		isPtr = true
		fullName =  xv.X.(*ast.Ident).Name
		name = fullName
	case *ast.Ident:
		isPtr = false
		fullName = xv.Name
		name = fullName
	case *ast.SelectorExpr:
		pkgName, _, isPtrr := getTypeName(xv.X, structs, pkg, imports)
		isPtr = isPtrr
		fullName = pkgName + "." + xv.Sel.Name
		name = fullName
		if imp, ok := imports[pkgName]; ok {
			fullName = imp.Path + "." + xv.Sel.Name
		}
	default:
		panic(fmt.Sprintf("no type found: %T", exp))
	}

	potentialFullName := fullTypeName(pkg, fullName)
	if _, ok := structs[potentialFullName]; ok {
		fullName = potentialFullName
	}

	return fullName, name, isPtr
}


func maybeNewReceiver(fn *ast.FuncDecl, structs StructStore, pkg *packages.Package) *Param {
	var rec *Param

	for _, f := range fn.Recv.List {
		fullTypeName, typeName, isPtr := getTypeName(f.Type, structs, pkg, nil)
		rec = &Param{
			Name: f.Names[0].Name,
			IsPtr: isPtr,
			TypeName: typeName,
			FullTypeName: fullTypeName,
		}
		break
	}

	return rec
}

func fullTypeName(pkg *packages.Package, typeName string) string {
	return strings.Join([]string{pkg.PkgPath, typeName}, ".")
}
