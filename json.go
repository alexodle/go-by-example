package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)


type SampleStruct struct {
	Rest map[string]interface{} `yaml:",inline"`
	NumberString  string `yaml:"number_str,omitempty" validate:"nonzero"`
}

type Sample2 struct {
	*SampleStruct `yaml:",inline" validate:"nonzero"`
	NumberString2 string `yaml:"number_str2,omitempty" validate:"nonzero"`
}

func main() {
	/*obj := map[string]string{"v1": "1010", "v2": "{{varname}}", "v3": "{{varname}}1234"}
	str, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(str))*/
	//func2(obj)
	readWrite()
}

func func2(obj interface{}) {
	str, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(str))
}

func getType() interface{} {
	return new(Sample2)
}


func readWrite() {
	obj := getType()

	yamlStr := `---
number_str: 01234
number_str2: 01234
another_str: '01234'
another_list:
  - one
  - two
another_obj:
  key1: val1
  key2:
    - one
    - two
`

	fmt.Printf("hihi.type: %T\n", obj)
	if err := yaml.Unmarshal([]byte(yamlStr), obj); err != nil {
		panic(err)
	}

	fmt.Println(obj)

	str, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(str))



	fmt.Println()
	fmt.Println()
}