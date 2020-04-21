package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	fs, err := New("base/path")
	require.Nil(t, err)
	require.Equal(t, "base/path", fs.path)
}

func TestFilesystem(t *testing.T) {
	t.Run("FindCapeByUsername", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "capes")
		if err != nil {
			panic(fmt.Errorf("cannot crete temp directory for tests: %w", err))
		}
		defer os.RemoveAll(dir)

		t.Run("exists cape", func(t *testing.T) {
			file, err := os.Create(path.Join(dir, "username.png"))
			if err != nil {
				panic(fmt.Errorf("cannot create temp skin for tests: %w", err))
			}
			defer os.Remove(file.Name())

			fs, _ := New(dir)
			cape, err := fs.FindCapeByUsername("username")
			require.Nil(t, err)
			require.NotNil(t, cape)
			capeFile, _ := cape.File.(*os.File)
			require.Equal(t, file.Name(), capeFile.Name())
		})

		t.Run("not exists cape", func(t *testing.T) {
			fs, _ := New(dir)
			cape, err := fs.FindCapeByUsername("username")
			require.Nil(t, err)
			require.Nil(t, cape)
		})

		t.Run("empty username", func(t *testing.T) {
			fs, _ := New(dir)
			cape, err := fs.FindCapeByUsername("")
			require.Nil(t, err)
			require.Nil(t, cape)
		})
	})
}
