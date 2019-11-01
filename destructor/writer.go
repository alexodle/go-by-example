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
		printf(w, "v0 := o.impl.%s\n", method.Field.Name)
		return []string{"v0"}
	} else if method.ReturnType != nil {
		var newVarNames []string
		for i, _ := range method.ReturnType {
			newVarNames = append(newVarNames, fmt.Sprintf("v%d", i))
		}
		printf(w, "%s := o.impl.%s(%s)\n", strings.Join(newVarNames, ", "), method.Name, strings.Join(varNames, ", "))
		return newVarNames
	}
	printf(w, "o.impl.%s(%s)\n", method.Name, strings.Join(varNames, ", "))
	return nil
}

func wrapArrayParam(w io.Writer, oldVarName, newVarName string, p *Param) {
	refItem := ""
	derefArray := ""

	newFuncName := fmt.Sprintf("New%s", p.Interface.Name)
	if p.Type.LocalPkgName() != "" {
		newFuncName = p.Type.LocalPkgName() + "." + newFuncName
	}

	wrap := func() {
		printf(w, "for _, v := range %s%s {\n", derefArray, oldVarName)
		printf(w, "%s%s = append(%s%s, %s(%sv))\n", derefArray, newVarName, derefArray, newVarName, newFuncName, refItem)
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(p.Type))

	if !p.Type.OriginalType.IsArrayTypePtr {
		refItem = "&"
	}

	if p.Type.IsPtr {
		derefArray = "*"
		printf(w, "if %s != nil {\n", oldVarName)
		wrap()
		printf(w, "}\n")
	} else {
		wrap()
	}
}

func wrapMapParam(w io.Writer, oldVarName, newVarName string, p *Param) {
	toPtrItem := ""
	derefMap := ""
	refMap := ""

	newFuncName := fmt.Sprintf("New%s", p.Interface.Name)
	if p.Type.MapValueType.LocalPkgName() != "" {
		newFuncName = p.Type.MapValueType.LocalPkgName() + "." + newFuncName
	}

	wrap := func() {
		printf(w, "%s = %s%s{}\n", newVarName, refMap, formatTypeWithoutLeadingPtr(p.Type))
		printf(w, "for k, v := range %s%s {\n", derefMap, oldVarName)
		if derefMap != "" {
			printf(w, "(%s%s)[k] = %s(%sv)\n", derefMap, newVarName, newFuncName, toPtrItem)
		} else {
			printf(w, "%s[k] = %s(%sv)\n", newVarName, newFuncName, toPtrItem)
		}
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(p.Type))

	if !p.Type.MapValueType.OriginalType.IsPtr {
		toPtrItem = "&"
	}

	if p.Type.IsPtr {
		derefMap = "*"
		refMap = "&"
		printf(w, "if %s != nil {\n", oldVarName)
		wrap()
		printf(w, "}\n")
	} else {
		wrap()
	}
}

func wrapParams(w io.Writer, method *Method, varNames []string) []string {
	for i, p := range method.ReturnType {
		if p.Interface != nil {
			oldVarName := varNames[i]
			newVarName := fmt.Sprintf("newv%d", i)
			varNames[i] = newVarName

			if p.Type.IsMap {
				wrapMapParam(w, oldVarName, newVarName, p)
			} else {
				if p.Type.IsArray {
					wrapArrayParam(w, oldVarName, newVarName, p)
				} else {
					ref := ""
					newFuncName := fmt.Sprintf("New%s", p.Interface.Name)
					if p.Type.LocalPkgName() != "" {
						newFuncName = p.Type.LocalPkgName() + "." + newFuncName
					}
					if !p.Type.OriginalType.IsPtr {
						ref = "&"
					}
					printf(w, "%s := %s(%s%s)\n", newVarName, newFuncName, ref, oldVarName)
				}
			}
		}
	}
	return varNames
}

func unwrapArrayParam(w io.Writer, oldVarName, newVarName string, t *Type) {
	derefItem := ""
	derefArray := ""

	unwrap := func() {
		printf(w, "for _, v := range %s%s {\n", derefArray, oldVarName)
		printf(w, "%s%s = append(%s%s, %sv.GetImpl())\n", derefArray, newVarName, derefArray, newVarName, derefItem)
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(t.OriginalType))

	if !t.OriginalType.IsArrayTypePtr {
		derefItem = "*"
	}

	if t.IsPtr {
		derefArray = "*"
		printf(w, "if %s != nil {\n", oldVarName)
		unwrap()
		printf(w, "}\n")
	} else {
		unwrap()
	}
}

func unwrapMapParam(w io.Writer, oldVarName, newVarName string, t *Type) {
	derefValue := ""
	derefMap := ""
	refMap := ""

	unwrap := func() {
		printf(w, "%s = %s%s{}\n", newVarName, refMap, formatTypeWithoutLeadingPtr(t.OriginalType))
		printf(w, "for k, v := range %s%s {\n", derefMap, oldVarName)
		if derefMap != "" {
			printf(w, "(%s%s)[k] = %sv.GetImpl()\n", derefMap, newVarName, derefValue)
		} else {
			printf(w, "%s[k] = %sv.GetImpl()\n", newVarName, derefValue)
		}
		printf(w, "}\n")
	}

	printf(w, "var %s %s\n", newVarName, formatType(t.OriginalType))

	if !t.OriginalType.IsArrayTypePtr {
		derefValue = "*"
	}

	if t.IsPtr {
		derefMap = "*"
		refMap = "&"
		printf(w, "if %s != nil {\n", oldVarName)
		unwrap()
		printf(w, "}\n")
	} else {
		unwrap()
	}
}

func unwrapParams(w io.Writer, params ParamsList) []string {
	var unwrappedVarNames []string
	for _, p := range params {
		unwrappedVarNames = append(unwrappedVarNames, p.Name)
	}
	for i, p := range params {
		if p.Interface != nil {
			oldVarName := unwrappedVarNames[i]
			newVarName := fmt.Sprintf("new%s", oldVarName)
			unwrappedVarNames[i] = newVarName
			deref := ""

			if p.Type.IsMap {
				unwrapMapParam(w, oldVarName, newVarName, p.Type)
			} else {
				if p.Type.IsArray {
					unwrapArrayParam(w, oldVarName, newVarName, p.Type)
				} else {
					if !p.Type.OriginalType.IsPtr {
						deref = "*"
					}
					printf(w, "%s := %s%s.GetImpl()\n", newVarName, deref, oldVarName)
				}
			}
		}
	}
	return unwrappedVarNames
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

func formatTypeWithoutLeadingPtr(t *Type) string {
	return formatTypeWithOptions(t, true)
}

func formatType(t *Type) string {
	return formatTypeWithOptions(t, false)
}

func formatTypeWithOptions(t *Type, ignoreLeadingPtr bool) string {
	var parts []string
	if !ignoreLeadingPtr && t.IsPtr {
		parts = append(parts, "*")
	}
	if t.IsArray {
		parts = append(parts, "[]")
		if t.IsArrayTypePtr {
			parts = append(parts, "*")
		}
	}
	if t.IsMap {
		parts = append(parts, "map[")
		parts = append(parts, formatType(t.MapKeyType))
		parts = append(parts, "]")
		parts = append(parts, formatType(t.MapValueType))
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
