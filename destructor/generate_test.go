package destructor

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Generate(t *testing.T) {
	require.NoError(t, os.RemoveAll("testdata/actualoutput"))

	GenerateWrappers("testdata/input", "testdata/actualoutput")
	defer func() {
		_ = os.RemoveAll("testdata/actualoutput")
	}()
	require.NoError(t, filepath.Walk("testdata/expectedoutput", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		actualOutputPath := strings.Replace(path, "testdata/expectedoutput", "testdata/actualoutput", 1)

		expectedBytes, err := ioutil.ReadFile(path)
		require.NoError(t, err)

		actualBytes, err := ioutil.ReadFile(actualOutputPath)
		require.NoError(t, err)

		require.Equal(t, string(expectedBytes), string(actualBytes))
		return nil
	}))
}
