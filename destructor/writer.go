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
		fmt.Printf("Formatting new file: %s\n", fullPath)
		cmd := exec.Command("goimports", "-w", fullPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Errorf("gofmt failure: %s", string(output)))
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
		printf(w, "GetImpl() *%s\n", iface.OriginalStructTypeName)
		for _, meth := range iface.Methods {
			if len(meth.ReturnType) > 0 {
				printf(w, "%s(%s) (%s)\n", meth.Name, formatParams(meth.Params), formatParams(meth.ReturnType))
			} else {
				printf(w, "%s(%s)\n", meth.Name, formatParams(meth.Params))
			}
		}
		printf(w, "}\n\n")
		writeWrapperStruct(w, iface)
	}
}

func writeWrapperStruct(w io.Writer, iface *Interface) {
	printf(w, "func New%s(impl *%s) %s {\n", iface.Name, iface.OriginalStructTypeName, iface.Name)
	printf(w, "return &%s{impl: impl}\n", iface.WrapperStruct.Name)
	printf(w, "}\n\n")

	printf(w, "type %s struct {\n", iface.WrapperStruct.Name)
	printf(w, "impl *%s\n", iface.OriginalStructTypeName)
	printf(w, "}\n\n")

	printf(w, "func (o *%s) GetImpl() *%s {\n", iface.WrapperStruct.Name, iface.OriginalStructTypeName)
	printf(w, "return o.impl\n")
	printf(w, "}\n\n")
	for _, method := range iface.WrapperStruct.PublicMethods {
		if method.ReturnType != nil {
			printf(w, "func (o *%s) %s(%s) (%s) {\n", iface.WrapperStruct.Name, method.Name, formatParams(method.Params), formatParams(method.ReturnType))
		} else {
			printf(w, "func (o *%s) %s(%s) {\n", iface.WrapperStruct.Name, method.Name, formatParams(method.Params))
		}

		newVarNames := unwrapParams(w, method.Params)
		returnVarNames := applyToImpl(w, method, newVarNames)
		if len(method.ReturnType) > 0 {
			returnVarNames = wrapParams(w, method, returnVarNames)
			printf(w, "return %s\n", strings.Join(returnVarNames, ", "))
		}

		printf(w, "}\n\n")
	}
}

func applyToImpl(w io.Writer, method *Method, varNames []string) []string {
	if method.IsFieldSetter {
		printf(w, "o.impl.%s = %s\n", method.Field.Name, varNames[0])
		return nil
	} else if method.IsFieldGetter {
		printf(w, "retval := o.impl.%s\n", method.Field.Name)
		return []string{"retval"}
	} else if method.ReturnType != nil {
		var newVarNames []string
		for i, _ := range method.ReturnType {
			newVarNames = append(newVarNames, fmt.Sprintf("retval%d", i))
		}
		printf(w, "%s := o.impl.%s(%s)\n", strings.Join(newVarNames, ", "), method.Name, strings.Join(varNames, ", "))
		return newVarNames
	}
	printf(w, "o.impl.%s(%s)\n", method.Name, strings.Join(varNames, ", "))
	return nil
}

func getVarNames(varBaseName string, i int) (string, string) {
	var oldVarName string
	if i == 0 {
		oldVarName = varBaseName
	} else {
		oldVarName = fmt.Sprintf("%s_%d", varBaseName, i)
	}
	newVarName := fmt.Sprintf("%s_%d", varBaseName, i+1)
	return oldVarName, newVarName
}

func unwrapArrayType(w io.Writer, oldT, newT *ArrayType, varBaseName string, i int) int {
	oldVarName, newVarName := getVarNames(varBaseName, i)
	derefArray := ""

	unwrap := func() {
		innerVarBaseName := "it"
		oldInnerVar, _ := getVarNames(innerVarBaseName, 0)

		printf(w, "for _, %s := range %s%s {\n", oldInnerVar, derefArray, oldVarName)
		vi := unwrapType(w, oldT.Type, newT.Type, innerVarBaseName, 0)
		newInnerVar, _ := getVarNames(innerVarBaseName, vi)

		printf(w, "%s%s = append(%s%s, %s)\n", derefArray, newVarName, derefArray, newVarName, newInnerVar)
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(oldT))
	if newT.IsPtr {
		derefArray = "*"
		printf(w, "if %s != nil {\n", oldVarName)
		unwrap()
		printf(w, "}\n")
	} else {
		unwrap()
	}

	return i + 1
}

func unwrapMapType(w io.Writer, oldT, newT *MapType, varBaseName string, i int) int {
	oldVarName, newVarName := getVarNames(varBaseName, i)
	derefMap := ""
	refMap := ""

	unwrap := func() {
		printf(w, "%s = %s%s{}\n", newVarName, refMap, formatTypeWithoutLeadingPtr(oldT))

		innerVarBaseName := "it"
		oldInnerVar, _ := getVarNames(innerVarBaseName, 0)

		printf(w, "for k, %s := range %s%s {\n", oldInnerVar, derefMap, oldVarName)
		vi := unwrapType(w, oldT.ValueType, newT.ValueType, innerVarBaseName, 0)
		newInnerVar, _ := getVarNames(innerVarBaseName, vi)

		if derefMap != "" {
			printf(w, "(%s%s)[k] = %s\n", derefMap, newVarName, newInnerVar)
		} else {
			printf(w, "%s[k] = %s\n", newVarName, newInnerVar)
		}
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(oldT))

	if newT.IsPtr {
		derefMap = "*"
		refMap = "&"
		printf(w, "if %s != nil {\n", oldVarName)
		unwrap()
		printf(w, "}\n")
	} else {
		unwrap()
	}

	return i + 1
}

func unwrapType(w io.Writer, oldT, newT Type, varBaseName string, i int) int {
	if _, ok := newT.(*ModeledType); ok {
		oldVarName, newVarName := getVarNames(varBaseName, i)
		if oldT.(*ModeledType).IsPtr {
			printf(w, "%s := %s.GetImpl()\n", newVarName, oldVarName)
		} else {
			printf(w, "%s := *%s.GetImpl()\n", newVarName, oldVarName)
		}
		return i + 1
	}

	switch tt := oldT.(type) {
	case *ArrayType:
		return unwrapArrayType(w, tt, newT.(*ArrayType), varBaseName, i)
	case *MapType:
		return unwrapMapType(w, tt, newT.(*MapType), varBaseName, i)
	default:
		panic(fmt.Errorf("unsupported type: %T->%T", newT, oldT))
	}
	return i
}

func unwrapParam(w io.Writer, p *Param) string {
	i := unwrapType(w, p.Type.OriginalType, p.Type.Type, p.Name, 0)
	newVarName, _ := getVarNames(p.Name, i)
	return newVarName
}

func unwrapParams(w io.Writer, params ParamsList) []string {
	var unwrappedVarNames []string
	for _, p := range params {
		unwrappedVarNames = append(unwrappedVarNames, p.Name)
	}
	for i, p := range params {
		if p.Type.OriginalType != nil {
			newVarName := unwrapParam(w, p)
			unwrappedVarNames[i] = newVarName
		}
	}
	return unwrappedVarNames
}

func wrapParam(w io.Writer, p *Param, paramName string) string {
	if paramName == "" {
		paramName = p.Name
	}
	i := wrapType(w, p.Type.OriginalType, p.Type.Type, paramName, 0)
	newVarName, _ := getVarNames(paramName, i)
	return newVarName
}

func wrapParams(w io.Writer, method *Method, varNames []string) []string {
	for i, p := range method.ReturnType {
		if p.Type.OriginalType != nil {
			varNames[i] = wrapParam(w, p, varNames[i])
		}
	}
	return varNames
}

func wrapType(w io.Writer, oldT, newT Type, varBaseName string, i int) int {
	if mt, ok := newT.(*ModeledType); ok {
		oldVarName, newVarName := getVarNames(varBaseName, i)
		if oldT.(*ModeledType).IsPtr {
			printf(w, "%s := %s(%s)\n", newVarName, mt.NewFuncNameForPkg, oldVarName)
		} else {
			printf(w, "%s := %s(&%s)\n", newVarName, mt.NewFuncNameForPkg, oldVarName)
		}
		return i + 1
	}

	switch tt := oldT.(type) {
	case *ArrayType:
		return wrapArrayType(w, tt, newT.(*ArrayType), varBaseName, i)
	case *MapType:
		return wrapMapType(w, tt, newT.(*MapType), varBaseName, i)
	default:
		panic(fmt.Errorf("unsupported type: %T->%T", oldT, newT))
	}
	return i
}

func wrapArrayType(w io.Writer, oldT, newT *ArrayType, varBaseName string, i int) int {
	oldVarName, newVarName := getVarNames(varBaseName, i)
	derefArray := ""

	wrap := func() {
		innerVarBaseName := "it"
		oldInnerVar, _ := getVarNames(innerVarBaseName, 0)

		printf(w, "for _, %s := range %s%s {\n", oldInnerVar, derefArray, oldVarName)
		vi := wrapType(w, oldT.Type, newT.Type, innerVarBaseName, 0)
		newInnerVar, _ := getVarNames(innerVarBaseName, vi)

		printf(w, "%s%s = append(%s%s, %s)\n", derefArray, newVarName, derefArray, newVarName, newInnerVar)
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(newT))

	if oldT.IsPtr {
		derefArray = "*"
		printf(w, "if %s != nil {\n", oldVarName)
		wrap()
		printf(w, "}\n")
	} else {
		wrap()
	}

	return i + 1
}

func wrapMapType(w io.Writer, oldT, newT *MapType, varBaseName string, i int) int {
	oldVarName, newVarName := getVarNames(varBaseName, i)
	derefMap := ""
	refMap := ""

	wrap := func() {
		printf(w, "%s = %s%s{}\n", newVarName, refMap, formatTypeWithoutLeadingPtr(newT))

		innerVarBaseName := "it"
		oldInnerVar, _ := getVarNames(innerVarBaseName, 0)

		printf(w, "for k, %s := range %s%s {\n", oldInnerVar, derefMap, oldVarName)
		vi := wrapType(w, oldT.ValueType, newT.ValueType, innerVarBaseName, 0)
		newInnerVar, _ := getVarNames(innerVarBaseName, vi)

		if derefMap != "" {
			printf(w, "(%s%s)[k] = %s\n", derefMap, newVarName, newInnerVar)
		} else {
			printf(w, "%s[k] = %s\n", newVarName, newInnerVar)
		}
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(newT))

	if oldT.IsPtr {
		derefMap = "*"
		refMap = "&"
		printf(w, "if %s != nil {\n", oldVarName)
		wrap()
		printf(w, "}\n")
	} else {
		wrap()
	}
	return i + 1
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

func formatTypeWithoutLeadingPtr(t Type) string {
	return formatTypeWithOptions(t, true)
}

func formatType(t Type) string {
	return formatTypeWithOptions(t, false)
}

func formatTypeWithOptions(t Type, ignoreLeadingPtr bool) string {
	switch tt := t.(type) {
	case *TopLevelType:
		return formatTypeWithOptions(tt.Type, ignoreLeadingPtr)
	case *ModeledType:
		if tt.IsPtr && !ignoreLeadingPtr {
			return "*" + tt.LocalNameForPkg
		} else {
			return tt.LocalNameForPkg
		}
	case *ArrayType:
		if tt.IsPtr && !ignoreLeadingPtr {
			return "*[]" + formatTypeWithOptions(tt.Type, false)
		} else {
			return "[]" + formatTypeWithOptions(tt.Type, false)
		}
	case *MapType:
		if tt.IsPtr && !ignoreLeadingPtr {
			return fmt.Sprintf("*map[%s]%s",
				formatTypeWithOptions(tt.KeyType, false),
				formatTypeWithOptions(tt.ValueType, false))
		} else {
			return fmt.Sprintf("map[%s]%s",
				formatTypeWithOptions(tt.KeyType, false),
				formatTypeWithOptions(tt.ValueType, false))
		}
	}
	panic(fmt.Errorf("unsupported type: %T", t))
}

func printf(w io.Writer, s string, args ...interface{}) {
	_, err := fmt.Fprintf(w, s, args...)
	if err != nil {
		panic(err)
	}
}
