package destructor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_convertLocalTypeForFile(t *testing.T) {
	origPkg := &Package{Name: "a", Path: "/input/a"}
	origStruct := &Struct{Name: "StructA", File: &File{Path: "/input/a/a.go", Package: origPkg}}

	newPkg := &Package{Name: "a", Path: "/output/a"}
	newFile := &File{Path: "/output/a/a.go", Package: newPkg}
	newIFace := &Interface{Name: "StructA", File: newFile, OriginalStruct: origStruct}
	ifaces := InterfaceStore{origStruct.FullName(): newIFace}

	structs := StructStore{origStruct.FullName(): origStruct}
	modeler := &modeler{structStore: structs, wrapperStore: ifaces, inputDir: "/input/", outputDir: "/output/"}

	newType, imports := modeler.convertTypeForFile(newFile,
		&TopLevelType{Type: &MapType{
			KeyType:   &BaseType{IsBuiltin: true, Name: "string"},
			ValueType: &BaseType{Name: "StructA", Package: origPkg},
		}})

	require.Equal(t, imports, ImportStore{
		"orig_a": &Import{ExplicitName: "orig_a", Path: "/input/a"},
	})
	require.Equal(t, newType, &TopLevelType{
		OriginalType: &MapType{
			KeyType:   &ModeledType{BaseType: BaseType{IsBuiltin: true, Name: "string"}, LocalNameForPkg: "string"},
			ValueType: &ModeledType{BaseType: BaseType{Name: "StructA", Package: origPkg}, LocalNameForPkg: "orig_a.StructA"},
		},
		Type: &MapType{
			KeyType: &ModeledType{BaseType: BaseType{IsBuiltin: true, Name: "string"}, LocalNameForPkg: "string"},
			ValueType: &ModeledType{
				BaseType:          BaseType{Name: "StructA", Package: newPkg},
				LocalNameForPkg:   "StructA",
				Interface:         newIFace,
				NewFuncNameForPkg: "NewStructA",
			},
		},
	})
}

func Test_convertNonlocalTypeForFile(t *testing.T) {
	origPkg := &Package{Name: "b", Path: "/input/b"}
	origType := &TopLevelType{Type: &BaseType{Name: "StructB", Package: origPkg}}
	origStruct := &Struct{Name: "StructB", File: &File{Path: "/input/b/b.go", Package: origPkg}}

	newPkg := &Package{Name: "b", Path: "/output/b"}
	newFile := &File{Path: "/output/b/b.go", Package: newPkg}
	newIFace := &Interface{Name: "StructB", File: newFile, OriginalStruct: origStruct}
	ifaces := InterfaceStore{origStruct.FullName(): newIFace}

	writeToFile := &File{Path: "/output/a/a.go", Package: &Package{"a", "/output/a"}}
	structs := StructStore{origStruct.FullName(): origStruct}
	modeler := &modeler{structStore: structs, wrapperStore: ifaces, inputDir: "/input/", outputDir: "/output/"}
	newType, imports := modeler.convertTypeForFile(writeToFile, origType)

	require.Equal(t, imports, ImportStore{
		"orig_b": &Import{ExplicitName: "orig_b", Path: "/input/b"},
		"b":      &Import{ExplicitName: "b", Path: "/output/b"},
	})
	require.Equal(t, newType, &TopLevelType{
		OriginalType: &ModeledType{BaseType: BaseType{Name: "StructB", Package: origPkg}, LocalNameForPkg: "orig_b.StructB"},
		Type: &ModeledType{
			BaseType:          BaseType{Name: "StructB", Package: newPkg},
			LocalNameForPkg:   "b.StructB",
			NewFuncNameForPkg: "b.NewStructB",
			Interface:         newIFace,
		},
	})
}

func Test_convertPtrs(t *testing.T) {
	origPkg := &Package{Name: "a", Path: "/input/a"}
	origStruct := &Struct{Name: "StructA", File: &File{Path: "/input/a/a.go", Package: origPkg}}

	newPkg := &Package{Name: "a", Path: "/output/a"}
	newFile := &File{Path: "/output/a/a.go", Package: newPkg}
	newIFace := &Interface{Name: "StructA", File: newFile, OriginalStruct: origStruct}
	ifaces := InterfaceStore{origStruct.FullName(): newIFace}

	structs := StructStore{origStruct.FullName(): origStruct}
	modeler := &modeler{structStore: structs, wrapperStore: ifaces, inputDir: "/input/", outputDir: "/output/"}

	newType, imports := modeler.convertTypeForFile(newFile,
		&TopLevelType{Type: &ArrayType{IsPtr: true, Type: &BaseType{
			Name:    "StructA",
			Package: origPkg,
			IsPtr:   true,
		}},
		})

	require.Equal(t, imports, ImportStore{
		"orig_a": &Import{ExplicitName: "orig_a", Path: "/input/a"},
	})
	require.Equal(t, newType, &TopLevelType{
		OriginalType: &ArrayType{IsPtr: true, Type: &ModeledType{
			LocalNameForPkg: "orig_a.StructA",
			BaseType: BaseType{
				Name:    "StructA",
				Package: origPkg,
				IsPtr:   true,
			},
		}},
		Type: &ArrayType{IsPtr: true, Type: &ModeledType{
			LocalNameForPkg:   "StructA",
			NewFuncNameForPkg: "NewStructA",
			Interface:         newIFace,
			BaseType: BaseType{
				Name:    "StructA",
				Package: newPkg,
			},
		}},
	})
}
