package destructor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}

func Parse(dirs ...string) StructStore {
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
	path := imp.Path.Value[1 : len(imp.Path.Value)-1]
	parts := strings.Split(path, "/")

	newImport := &Import{Path: path, ImplicitName: parts[len(parts)-1]}
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

	fillStructs(f, directory, structs, &File{Path: filepath, Imports: createImportStore(f), Package: getPkg(directory)})
}

func isPublic(name string) bool {
	return strings.ToUpper(name[0:1]) == name[0:1]
}

func fillMethods(node ast.Node, directory string, structs StructStore, store ImportStore) {
	pkg := getPkg(directory)

	ast.Walk(walker(func(node ast.Node) bool {
		switch v := node.(type) {
		case *ast.FuncDecl:
			m := newMethod(v, structs, pkg, store)
			if m.Receiver != nil && isPublic(m.Name) {
				s := structs[m.Receiver.Type.FullName]
				s.PublicMethods = append(s.PublicMethods, m)
			}
		case *ast.TypeSpec:
			if t, ok := v.Type.(*ast.StructType); ok {
				structName := fullTypeName(pkg, v.Name.Name)
				fields := newFields(t, structs, pkg, store)
				structs[structName].Fields = fields
			}
		}
		return true
	}), node)
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
		File:          file,
		Name:          v.Name.Name,
		FullName:      fullTypeName(file.Package, v.Name.Name),
		PublicMethods: MethodList{},
	}
}

func newFields(v *ast.StructType, structs StructStore, pkg *Package, imports ImportStore) ParamsList {
	var fields ParamsList
	if v.Fields != nil {
		for _, f := range v.Fields.List {
			t := getTypeName(f.Type, structs, pkg, imports)
			if f.Names != nil {
				for _, pn := range f.Names {
					if isPublic(pn.Name) {
						fields = append(fields, &Param{Name: pn.Name, Type: t})
					}
				}
			} else if isPublic(t.Name) {
				fields = append(fields, &Param{Name: t.Name, Type: t})
			}
		}
	}
	return fields
}

func newMethod(v *ast.FuncDecl, structs StructStore, pkg *Package, imports ImportStore) *Method {
	return &Method{
		Name:       v.Name.Name,
		Receiver:   maybeNewReceiver(v, structs, pkg),
		Params:     getParams(v.Type.Params, structs, pkg, imports),
		ReturnType: getReturnParams(v.Type.Results, structs, pkg, imports),
	}
}

func getReturnParams(fieldList *ast.FieldList, structs StructStore, pkg *Package, imports ImportStore) ParamsList {
	if fieldList == nil {
		return nil
	}
	var params []*Param
	for _, p := range fieldList.List {
		t := getTypeName(p.Type, structs, pkg, imports)
		if p.Names != nil {
			for _, pn := range p.Names {
				params = append(params, &Param{Name: pn.Name, Type: t})
			}
		} else {
			params = append(params, &Param{Type: t})
		}
	}
	return params
}

func getParams(fieldList *ast.FieldList, structs StructStore, pkg *Package, imports ImportStore) ParamsList {
	if fieldList == nil {
		return nil
	}
	var params []*Param
	for _, p := range fieldList.List {
		t := getTypeName(p.Type, structs, pkg, imports)
		for _, pn := range p.Names {
			params = append(params, &Param{Name: pn.Name, Type: t})
		}
	}
	return params
}

func getTypeName(exp ast.Expr, structs StructStore, pkg *Package, imports ImportStore) *Type {
	var fullName string
	var name string
	switch xv := exp.(type) {
	case *ast.Ident:
		fullName = xv.Name
		name = fullName
	case *ast.StarExpr:
		tn := getTypeName(xv.X, structs, pkg, imports)
		tn.IsPtr = true
		return tn
	case *ast.SelectorExpr:
		tn := getTypeName(xv.X, structs, pkg, imports)
		fullName = tn.FullName + "." + xv.Sel.Name
		name = fullName
		if imp, ok := imports[tn.FullName]; ok {
			fullName = imp.Path + "." + xv.Sel.Name
		}
	case *ast.ArrayType:
		tn := getTypeName(xv.Elt, structs, pkg, imports)
		return &Type{
			FullName:       tn.FullName,
			Name:           tn.Name,
			IsArray:        true,
			IsArrayTypePtr: tn.IsPtr,
		}
	case *ast.MapType:
		ktype := getTypeName(xv.Key, structs, pkg, imports)
		vtype := getTypeName(xv.Value, structs, pkg, imports)
		return &Type{
			IsMap:        true,
			MapKeyType:   ktype,
			MapValueType: vtype,
		}
	case *ast.FuncType:
		return &Type{
			IsFunc:         true,
			FuncParams:     getParams(xv.Params, structs, pkg, imports),
			FuncReturnType: getReturnParams(xv.Results, structs, pkg, imports),
		}
	case *ast.InterfaceType:
		if xv.Methods != nil && xv.Methods.List != nil && len(xv.Methods.List) > 0 {
			panic(fmt.Errorf("non-empty interface params not supported"))
		}
		return &Type{
			Name:     "interface{}",
			FullName: "interface{}",
		}
	default:
		panic(fmt.Sprintf("no type found: %T", exp))
	}

	potentialFullName := fullTypeName(pkg, fullName)
	if _, ok := structs[potentialFullName]; ok {
		fullName = potentialFullName
	}

	return &Type{FullName: fullName, Name: name}
}

func maybeNewReceiver(fn *ast.FuncDecl, structs StructStore, pkg *Package) *Param {
	var rec *Param

	if fn.Recv != nil {
		for _, f := range fn.Recv.List {
			t := getTypeName(f.Type, structs, pkg, nil)
			rec = &Param{
				Name: f.Names[0].Name,
				Type: t,
			}
			break
		}
	}

	return rec
}

func fullTypeName(pkg *Package, typeName string) string {
	return strings.Join([]string{pkg.Path, typeName}, ".")
}

func getPkg(directory string) *Package {
	fullPath, err := filepath.Abs(directory)
	if err != nil {
		panic(err)
	}
	gopath := os.Getenv("GOPATH")
	fullPath = strings.TrimPrefix(fullPath, gopath+"/src/")
	pkgs, err := packages.Load(nil, fullPath)
	if err != nil {
		panic(err)
	} else if len(pkgs) != 1 {
		panic(fmt.Sprintf("found multiple packages for directory: %s", directory))
	}
	pkg := pkgs[0]
	if pkg.Errors != nil {
		panic(pkg.Errors[0])
	}
	return &Package{
		Path: pkg.PkgPath,
		Name: pkg.Name,
	}
}
