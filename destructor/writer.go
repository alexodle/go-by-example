package destructor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func WriteCode(files []*File) {
	for _, f := range files {
		writeFile(f)
	}
	for _, f := range files {
		fullPath, err := filepath.Abs(f.Path)
		if err != nil {
			panic(err)
		}
		cmd := exec.Command("go", "fmt", fullPath)
		_, err = cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
	}
}

func writeFile(f *File) {
	fullPath, err := filepath.Abs(f.Path)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Writing new file: %s\n", fullPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		panic(err)
	}
	w, err := os.Create(fullPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = w.Close()
		if err != nil {
			panic(err)
		}
	}()

	printf(w, "package %s\n\n", f.Package.Name)

	writeImports(w, f.Imports)
	printf(w, "\n")

	writeInterfaces(w, f.Interfaces)
	printf(w, "\n")
}

func writeImports(w io.Writer, imps ImportStore) {
	var impStrs []string
	for _, imp := range imps {
		impStrs = append(impStrs, fmt.Sprintf("import %s \"%s\"\n", imp.ExplicitName, imp.Path))
	}
	sort.Strings(impStrs)
	for _, s := range impStrs {
		printf(w, s)
	}
}

func writeInterfaces(w io.Writer, interfaces InterfaceList) {
	for _, iface := range interfaces {
		printf(w, "type %s interface {\n", iface.Name)
		printf(w, "\tGetImpl() *%s\n", iface.OriginalStructTypeName)
		for _, meth := range iface.Methods {
			if len(meth.ReturnType) > 0 {
				printf(w, "\t%s(%s) (%s)\n", meth.Name, formatParams(meth.Params), formatParams(meth.ReturnType))
			} else {
				printf(w, "\t%s(%s)\n", meth.Name, formatParams(meth.Params))
			}
		}
		printf(w, "}\n\n")
		writeWrapperStruct(w, iface)
	}
}

func writeWrapperStruct(w io.Writer, iface *Interface) {
	printf(w, "func New%s(impl *%s) %s {\n", iface.Name, iface.OriginalStructTypeName, iface.Name)
	printf(w, "\treturn &%s{impl: impl}\n", iface.WrapperStruct.Name)
	printf(w, "}\n\n")

	printf(w, "type %s struct {\n", iface.WrapperStruct.Name)
	printf(w, "\timpl *%s\n", iface.OriginalStructTypeName)
	printf(w, "}\n\n")

	printf(w, "func (o *%s) GetImpl() *%s {\n", iface.WrapperStruct.Name, iface.OriginalStructTypeName)
	printf(w, "\treturn o.impl\n")
	printf(w, "}\n\n")
	for _, method := range iface.WrapperStruct.PublicMethods {
		if method.ReturnType != nil {
			printf(w, "func (o *%s) %s(%s) (%s) {\n", iface.WrapperStruct.Name, method.Name, formatParams(method.Params), formatParams(method.ReturnType))
			writeFunctionImpl(w, method)
		} else {
			printf(w, "func (o *%s) %s(%s) {\n", iface.WrapperStruct.Name, method.Name, formatParams(method.Params))
			if method.IsFieldSetter {
				printf(w, "\to.impl.%s = %s\n", method.Field.Name, method.Params[0].Name)
			} else {
				printf(w, "\to.impl.%s(%s)\n", method.Name, formatParamsCall(method.Params))
			}
		}
		printf(w, "}\n\n")
	}
}

func formatParamsCall(params ParamsList) string {
	var strs []string
	for _, p := range params {
		if p.Interface != nil {
			if p.Type.OriginalType.IsPtr {
				strs = append(strs, fmt.Sprintf("%s.GetImpl()", p.Name))
			} else {
				strs = append(strs, fmt.Sprintf("*%s.GetImpl()", p.Name))
			}
		} else {
			strs = append(strs, p.Name)
		}
	}
	return strings.Join(strs, ", ")
}

func writeFunctionImpl(w io.Writer, method *Method) {
	var localVarNames []string
	for i, _ := range method.ReturnType {
		localVarNames = append(localVarNames, fmt.Sprintf("v%d", i))
	}

	if method.IsFieldGetter {
		printf(w, "\t%s := o.impl.%s\n", strings.Join(localVarNames, ", "), method.Field.Name)
	} else {
		printf(w, "\t%s := o.impl.%s(%s)\n", strings.Join(localVarNames, ", "), method.Name, formatParamsCall(method.Params))
	}

	for i, p := range method.ReturnType {
		if p.Interface != nil {
			oldVarName := localVarNames[i]
			newVarName := fmt.Sprintf("newv%d", i)
			localVarNames[i] = newVarName
			toPtr := ""

			if p.Type.IsMap {
				if !p.Type.MapValueType.OriginalType.IsPtr {
					toPtr = "&"
				}
				newFuncName := fmt.Sprintf("New%s", p.Interface.Name)
				if p.Type.MapValueType.LocalPkgName() != "" {
					newFuncName = p.Type.MapValueType.LocalPkgName() + "." + newFuncName
				}
				printf(w, "\tvar %s %s\n", newVarName, formatType(p.Type))
				printf(w, "\tfor k, v := range %s {\n", oldVarName)
				printf(w, "\t\t%s[k] = %s(%sv)\n", newVarName, newFuncName, toPtr)
				printf(w, "\t}\n")
			} else {
				newFuncName := fmt.Sprintf("New%s", p.Interface.Name)
				if p.Type.LocalPkgName() != "" {
					newFuncName = p.Type.LocalPkgName() + "." + newFuncName
				}
				if p.Type.IsArray {
					if !p.Type.OriginalType.IsArrayTypePtr {
						toPtr = "&"
					}
					printf(w, "\tvar %s %s\n", newVarName, formatType(p.Type))
					printf(w, "\tfor _, v := range %s {\n", oldVarName)
					printf(w, "\t\t%s = append(%s, %s(%sv))\n", newVarName, newVarName, newFuncName, toPtr)
					printf(w, "\t}\n")
				} else {
					if !p.Type.OriginalType.IsPtr {
						toPtr = "&"
					}
					printf(w, "\t%s := %s(%s%s)\n", newVarName, newFuncName, toPtr, oldVarName)
				}
			}
		}
	}

	printf(w, "\treturn %s\n", strings.Join(localVarNames, ", "))
}

func formatParams(params ParamsList) string {
	var strs []string
	for _, p := range params {
		typeStr := formatType(p.Type)
		if p.Name != "" {
			strs = append(strs, fmt.Sprintf("%s %s", p.Name, typeStr))
		} else {
			strs = append(strs, typeStr)
		}
	}
	return strings.Join(strs, ", ")
}

func formatType(t *Type) string {
	var parts []string
	if t.IsPtr {
		parts = append(parts, "*")
	}
	if t.IsArray {
		parts = append(parts, "[]")
		if t.IsArrayTypePtr {
			parts = append(parts, "*")
		}
	} else if t.IsMap {
		parts = append(parts, "map[")
		parts = append(parts, formatType(t.MapKeyType))
		parts = append(parts, "]")
		parts = append(parts, formatType(t.MapValueType))
		return strings.Join(parts, "")
	} else if t.IsEmptyInterface {
		parts = append(parts, "interface{}")
		return strings.Join(parts, "")
	}

	parts = append(parts, t.Name)
	return strings.Join(parts, "")
}

func printf(w io.Writer, s string, args ...interface{}) {
	_, err := fmt.Fprintf(w, s, args...)
	if err != nil {
		panic(err)
	}
}
