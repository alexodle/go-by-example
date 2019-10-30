package destructor

type StructStore map[string]*Struct
type ImportStore map[string]*Import
type InterfaceStore map[string]*Interface

type ParamsList []*Param
type MethodList []*Method
type InterfaceList []*Interface

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
	FullName               string
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
	File          *File
	Name          string
	FullName      string
	PublicMethods MethodList
	Fields        ParamsList
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
	Name      string
	Type      *Type
	Interface *Interface
}

type Type struct {
	FullName       string
	Name           string
	IsPtr          bool
	IsArray        bool
	IsArrayTypePtr bool
}
