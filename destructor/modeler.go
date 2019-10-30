package destructor

import (
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

type modeler struct{
	structStore StructStore
	wrapperStore InterfaceStore
	inputDir string
	outputDir string
}

func shouldWrap(st *Struct) bool {
	return len(st.PublicMethods) > 0
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
				Path:    newPath,
				Imports: ImportStore{
					"orig_"+st.File.Package.Name: &Import{ExplicitName: "orig_" + st.File.Package.Name, Path: st.File.Package.Path},
				},
				Package: &Package{
					Name: st.File.Package.Name,
					Path: strings.Replace(st.File.Package.Path, m.inputDir, m.outputDir, 1),
				},
			}
			newFiles[newPath] = newFile
		}

		newIFace := &Interface{
			File: newFile,
			Name: st.Name,
			FullName: newFile.Package.Path + "." + st.Name,
			Methods: MethodList{},
			OriginalStruct: st,
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
			File: iface.File,
			Name: newStructName,
			FullName: newStructName,
			PublicMethods: MethodList{},
		}

		fields, fieldImps := m.convertTypesForFile(iface.File, iface.OriginalStruct.Fields)
		for _, f := range fields {
			setParams := ParamsList{&Param{Name: "v", Type: f.Type}}
			getReturnType := ParamsList{&Param{Type: f.Type}}
			iface.Methods = append(iface.Methods,
				&Method{
					Name: "Get"+f.Name,
					ReturnType: getReturnType,
				},
				&Method{
					Name: "Set"+f.Name,
					Params: setParams,
				},
			)
			iface.WrapperStruct.PublicMethods = append(iface.WrapperStruct.PublicMethods,
				&Method{
					Name: "Get"+f.Name,
					Receiver: recvParam,
					ReturnType: getReturnType,
					IsFieldGetter: true,
					Field: f,
				},
				&Method{
					Name: "Set"+f.Name,
					Receiver: recvParam,
					Params: setParams,
					IsFieldSetter: true,
					Field: f,
				},
			)
		}
		mergeImportStores(iface.File.Imports, fieldImps)

		for _, method := range iface.OriginalStruct.PublicMethods {
			params, imps1 := m.convertTypesForFile(iface.File, method.Params)
			returnType, imps2 := m.convertTypesForFile(iface.File, method.ReturnType)
			iface.Methods = append(iface.Methods, &Method{
				Name: method.Name,
				Params: params,
				ReturnType: returnType,
			})
			iface.WrapperStruct.PublicMethods = append(iface.WrapperStruct.PublicMethods, &Method{
				Name: method.Name,
				Params: params,
				ReturnType: returnType,
				Receiver: recvParam,
			})
			mergeImportStores(iface.File.Imports, imps1)
			mergeImportStores(iface.File.Imports, imps2)
		}
	}
}

func (m *modeler) convertTypesForFile(f *File, params ParamsList) (ParamsList, ImportStore) {
	var newParams ParamsList
	importStore := ImportStore{}
	for _, p := range params {
		t, imp, iface := m.convertTypeForFile(f, p.Type)
		newParams = append(newParams, &Param{
			Name: p.Name,
			Type: t,
			Interface: iface,
		})
		if imp != nil {
			importStore[imp.ExplicitName] = imp
		}
	}
	return newParams, importStore
}

func (m *modeler) convertTypeForFile(f *File, t *Type) (*Type, *Import, *Interface) {
	if t.Name == t.FullName {
		return t, nil, nil
	}

	var newType = *t // shallow copy
	newType.FullName = ""

	var iface *Interface
	fullTypeName := t.FullName
	prefix := "orig_"
	if wrapper, ok := m.wrapperStore[fullTypeName]; ok {
		if wrapper.File.Package.Path == f.Package.Path {
			newType.Name = wrapper.Name
			return &newType, nil, wrapper
		}
		iface = wrapper
		prefix = ""
		fullTypeName = wrapper.FullName
	}

	parts := strings.Split(fullTypeName, "/")
	typeNameStr := parts[len(parts) - 1]
	importName := strings.Split(typeNameStr, ".")[0]
	importPath := strings.Join(parts[:len(parts) - 1], "/") + "/" + importName

	newType.Name = prefix + typeNameStr
	return &newType, &Import{ ExplicitName: prefix + importName, Path: importPath}, iface
}

func mergeImportStores(a ImportStore, b ImportStore) {
	for k, b := range b {
		a[k] = b
	}
}
