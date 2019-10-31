package destructor

import (
	"fmt"
	"strings"
)

func Remodel(structs StructStore, inputDir, outputDir string) []*File {
	modeler := &modeler{structStore: structs, wrapperStore: InterfaceStore{}, inputDir: inputDir, outputDir: outputDir}
	modeler.buildWrappers()
	modeler.fillWrappers()

	seenFiles := map[string]struct{}{}
	var filesList []*File
	for _, iface := range modeler.wrapperStore {
		if _, ok := seenFiles[iface.File.Path]; ok {
			continue
		}
		filesList = append(filesList, iface.File)
		seenFiles[iface.File.Path] = struct{}{}
	}
	return filesList
}

type modeler struct {
	structStore  StructStore
	wrapperStore InterfaceStore
	inputDir     string
	outputDir    string
}

func shouldWrap(st *Struct) bool {
	return len(st.PublicMethods) > 0 || len(st.Fields) > 0
}

func (m *modeler) buildWrappers() {
	newFiles := map[string]*File{}
	for key, st := range m.structStore {
		if !shouldWrap(st) {
			continue
		}

		newPath := strings.Replace(st.File.Path, m.inputDir, m.outputDir, 1)
		newFile, ok := newFiles[newPath]
		if !ok {
			newFile = &File{
				Path: newPath,
				Imports: ImportStore{
					"orig_" + st.File.Package.Name: &Import{ExplicitName: "orig_" + st.File.Package.Name, Path: st.File.Package.Path},
				},
				Package: &Package{
					Name: st.File.Package.Name,
					Path: strings.Replace(st.File.Package.Path, m.inputDir, m.outputDir, 1),
				},
			}
			newFiles[newPath] = newFile
		}

		newIFace := &Interface{
			File:                   newFile,
			Name:                   st.Name,
			FullName:               newFile.Package.Path + "." + st.Name,
			Methods:                MethodList{},
			OriginalStruct:         st,
			OriginalStructTypeName: "orig_" + st.File.Package.Name + "." + st.Name,
		}
		m.wrapperStore[key] = newIFace
		newFile.Interfaces = append(newFile.Interfaces, newIFace)
	}
}

func (m *modeler) fillWrappers() {
	for _, iface := range m.wrapperStore {
		newStructName := strings.ToLower(iface.OriginalStruct.Name[0:1]) + iface.OriginalStruct.Name[1:] + "Wrapper"
		recvParam := &Param{Name: "o", Type: &Type{IsPtr: true, Name: newStructName}}

		iface.WrapperStruct = &Struct{
			File:          iface.File,
			Name:          newStructName,
			FullName:      newStructName,
			PublicMethods: MethodList{},
		}

		fields, fieldImps := m.convertTypesForFile(iface.File, iface.OriginalStruct.Fields)
		iface.File.Imports.AddAll(fieldImps)

		for _, f := range fields {
			if isUnsupportedType(f.Type) {
				fmt.Printf("WARN: skipping getter/setter methods for field:%s.%s, at least one param or return type is not currently supported\n", iface.Name, f.Name)
				continue
			}
			setParams := ParamsList{&Param{Name: "v", Type: f.Type, Interface: f.Interface}}
			getReturnType := ParamsList{&Param{Type: f.Type, Interface: f.Interface}}
			iface.Methods = append(iface.Methods,
				&Method{
					Name:       "Get" + f.Name,
					ReturnType: getReturnType,
				},
				&Method{
					Name:   "Set" + f.Name,
					Params: setParams,
				},
			)
			iface.WrapperStruct.PublicMethods = append(iface.WrapperStruct.PublicMethods,
				&Method{
					Name:          "Get" + f.Name,
					Receiver:      recvParam,
					ReturnType:    getReturnType,
					IsFieldGetter: true,
					Field:         f,
				},
				&Method{
					Name:          "Set" + f.Name,
					Receiver:      recvParam,
					Params:        setParams,
					IsFieldSetter: true,
					Field:         f,
				},
			)
		}

		for _, method := range iface.OriginalStruct.PublicMethods {
			if isUnsupportedMethod(method) {
				fmt.Printf("WARN: skipping method:%s.%s, at least one param or return type is not currently supported\n", iface.Name, method.Name)
				continue
			}

			params, imps1 := m.convertTypesForFile(iface.File, method.Params)
			iface.File.Imports.AddAll(imps1)

			returnType, imps2 := m.convertTypesForFile(iface.File, method.ReturnType)
			iface.File.Imports.AddAll(imps2)

			iface.Methods = append(iface.Methods, &Method{
				Name:       method.Name,
				Params:     params,
				ReturnType: returnType,
			})
			iface.WrapperStruct.PublicMethods = append(iface.WrapperStruct.PublicMethods, &Method{
				Name:       method.Name,
				Params:     params,
				ReturnType: returnType,
				Receiver:   recvParam,
			})
		}
	}
}

func (m *modeler) convertTypesForFile(f *File, params ParamsList) (ParamsList, ImportStore) {
	var newParams ParamsList
	importStore := ImportStore{}
	for _, p := range params {
		t, imps, iface := m.convertTypeForFile(f, p.Type)
		newParams = append(newParams, &Param{
			Name:      p.Name,
			Type:      t,
			Interface: iface,
		})
		importStore.AddAll(imps)
	}
	return newParams, importStore
}

func (m *modeler) convertMapTypeForFile(f *File, t *Type) (*Type, ImportStore, *Interface) {
	// shallow copy
	var newType = *t
	newType.OriginalType = t

	keyType, imports1, key_iface := m.convertTypeForFile(f, t.MapKeyType)
	if key_iface != nil {
		panic(fmt.Errorf("struct map keys not supported yet: %s", key_iface.OriginalStruct.FullName))
	}
	newType.MapKeyType = keyType

	valType, imports2, iface := m.convertTypeForFile(f, t.MapValueType)
	newType.MapValueType = valType

	imports1.AddAll(imports2)
	return &newType, imports1, iface
}

func (m *modeler) convertTypeForFile(f *File, t *Type) (*Type, ImportStore, *Interface) {
	if t.IsMap {
		return m.convertMapTypeForFile(f, t)
	}

	if !strings.Contains(t.FullName, ".") {
		return t, ImportStore{}, nil
	}

	imports := ImportStore{}

	var newType = *t // shallow copy
	newType.OriginalType = t
	newType.FullName = ""

	var iface *Interface
	fullTypeName := t.FullName
	prefix := "orig_"
	if wrapper, ok := m.wrapperStore[fullTypeName]; ok {
		newType.OriginalType.Name = "orig_" + newType.OriginalType.Name
		addImportByFullName(newType.OriginalType.FullName, imports, "orig_")

		if newType.IsArray {
			newType.IsArrayTypePtr = false
		} else {
			newType.IsPtr = false
		}

		if wrapper.File.Package.Path == f.Package.Path {
			newType.Name = wrapper.Name
			return &newType, imports, wrapper
		}

		iface = wrapper
		prefix = ""
		fullTypeName = wrapper.FullName
	}

	parts := strings.Split(fullTypeName, "/")
	newType.Name = prefix + parts[len(parts)-1]
	addImportByFullName(fullTypeName, imports, prefix)
	return &newType, imports, iface
}

func addImportByFullName(tn string, imports ImportStore, namePrefix string) {
	name, path, ok := extractImportFromFullPath(tn)
	if !ok {
		panic(fmt.Errorf("failed to parse fullname: %s", tn))
	}
	imports[namePrefix+name] = &Import{ExplicitName: namePrefix + name, Path: path}
}

func extractImportFromFullPath(tn string) (string, string, bool) {
	slashSplits := strings.Split(tn, "/")
	nameParts := slashSplits[len(slashSplits)-1]
	typeName := strings.Split(nameParts, ".")
	if len(typeName) > 1 {
		importPath := strings.TrimSuffix(tn, "."+typeName[1])
		return typeName[0], importPath, true
	}

	return "", "", false
}

func isUnsupportedType(t *Type) bool {
	return t.IsFunc
}

func hasUnsupportedType(params ParamsList) bool {
	for _, p := range params {
		if isUnsupportedType(p.Type) {
			return true
		}
	}
	return false
}

func isUnsupportedMethod(method *Method) bool {
	return hasUnsupportedType(method.Params) || hasUnsupportedType(method.ReturnType)
}
