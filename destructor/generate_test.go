package destructor

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_Generate(t *testing.T) {
	require.NoError(t, os.RemoveAll("testdata/generated"))
	GenerateWrappers("testdata/input", "testdata/generated")
}

func Test_GeneratedCode(t *testing.T) {
}
