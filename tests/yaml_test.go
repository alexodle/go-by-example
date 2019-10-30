package main

import (
	"fmt"
	"os"
	"testing"
	"gopkg.in/yaml.v2"
	"github.com/stretchr/testify/require"
)

type X struct {
	X string `yaml:"x"`
}

func TestThing(t *testing.T) {
	var x X
	err:=yaml.Unmarshal([]byte("x: 010"),&x)
	require.NoError(t,err)
	fmt.Printf("%#v\n",x)
}

func TestThing2(t *testing.T) {
	var x1 interface{}
	err:=yaml.Unmarshal([]byte("x: 010"),&x1)
	require.NoError(t,err)
	fmt.Printf("%#v\n",x1)
	x1bytes, err:=yaml.Marshal(x1)
	require.NoError(t,err)
	var x2 X
	yaml.Unmarshal(x1bytes,&x2)
	fmt.Printf("%#v\n",x2)
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}