package destructor

import "sort"

type StructStore map[string]*Struct
type InterfaceStore map[string]*Interface
type ImportStore map[string]*Import

func (i *ImportStore) AddAll(other ImportStore) {
	for k, imp := range other {
		(*i)[k] = imp
	}
}

func (i *ImportStore) ToSortedList() ImportList {
	var l ImportList
	for _, v := range *i {
		l = append(l, v)
	}
	sort.Sort(l)
	return l
}

type ParamsList []*Param
type MethodList []*Method
type InterfaceList []*Interface
type ImportList []*Import

type Import struct {
	ImplicitName string
	ExplicitName string
	Path         string
}

type Package struct {
	Name string
	Path string
}

type Interface struct {
	File                   *File
	Name                   string
	Methods                MethodList
	OriginalStruct         *Struct
	OriginalStructTypeName string
	WrapperStruct          *Struct
}

type File struct {
	Path       string
	Imports    ImportStore
	Package    *Package
	Interfaces InterfaceList
}

type Struct struct {
	Name          string
	File          *File
	PublicMethods MethodList
	Fields        ParamsList
}

func (s *Struct) FullName() string {
	return s.File.Package.Path + "." + s.Name
}

type Method struct {
	Name          string
	Receiver      *Param
	Params        ParamsList
	ReturnType    ParamsList
	IsFieldSetter bool
	IsFieldGetter bool
	Field         *Param
}

type Param struct {
	Name string
	Type *TopLevelType
}

// Sorting

func (l ImportList) Len() int {
	return len(l)
}
func (l ImportList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l ImportList) Less(i, j int) bool {
	return l[i].ExplicitName < l[j].ExplicitName
}

func (l InterfaceList) Len() int {
	return len(l)
}
func (l InterfaceList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l InterfaceList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l MethodList) Len() int {
	return len(l)
}
func (l MethodList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l MethodList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

// Types

type Type interface {
	ShallowCopy() Type
	GetBaseType() Type
}

type TopLevelType struct {
	OriginalType Type
	Type         Type
}

func (t *TopLevelType) ShallowCopy() Type {
	t2 := *t
	return &t2
}

func (t *TopLevelType) GetBaseType() Type {
	return t.Type.GetBaseType()
}

type BaseType struct {
	Name      string
	IsBuiltin bool
	Package   *Package
	IsBuiltIn bool
	IsPtr     bool
}

func (t *BaseType) GetBaseType() Type {
	return t
}

func (t *BaseType) FullName() string {
	if t.IsBuiltin {
		return t.Name
	}
	return t.Package.Path + "." + t.Name
}

func (t *BaseType) ShallowCopy() Type {
	t2 := *t
	return &t2
}

type ModeledType struct {
	BaseType
	LocalNameForPkg   string
	NewFuncNameForPkg string
	Interface         *Interface
}

type ArrayType struct {
	Type  Type
	IsPtr bool
}

func (t *ArrayType) GetBaseType() Type {
	return t.Type.GetBaseType()
}

func (t *ArrayType) ShallowCopy() Type {
	t2 := *t
	t2.Type = t2.Type.ShallowCopy()
	return &t2
}

type MapType struct {
	KeyType   Type
	ValueType Type
	IsPtr     bool
}

func (t *MapType) GetBaseType() Type {
	return t.ValueType.GetBaseType()
}

func (t *MapType) ShallowCopy() Type {
	t2 := *t
	t2.ValueType = t2.ValueType.ShallowCopy()
	t2.KeyType = t2.KeyType.ShallowCopy()
	return &t2
}

type FuncType struct {
	FuncParams     ParamsList
	FuncReturnType ParamsList
}

func (t *FuncType) GetBaseType() Type {
	return t
}

func (t *FuncType) ShallowCopy() Type {
	t2 := *t
	return &t2
}
