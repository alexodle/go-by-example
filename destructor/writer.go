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
		cmd := exec.Command("go", "fmt", f.Path)
		_, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
	}
}

func writeFile(f *File) {
	if err := os.MkdirAll(filepath.Dir(f.Path), os.ModePerm); err != nil {
		panic(err)
	}
	w, err := os.Create(f.Path)
	if err != nil {
		panic(err)
	}
	defer func() { _ = w.Close() }()

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
			if method.IsFieldGetter {
				printf(w, "\treturn o.impl.%s\n", method.Field.Name)
			} else {
				printf(w, "\treturn o.impl.%s(%s)\n", method.Name, formatParamsCall(method.Params))
			}
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
			if p.Type.IsPtr {
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
	}
	parts = append(parts, t.Name)
	return strings.Join(parts, "")
}

func printf(w io.Writer, s string, args... interface{}) {
	_, err := fmt.Fprintf(w, s, args...)
	if err != nil {
		panic(err)
	}
}
