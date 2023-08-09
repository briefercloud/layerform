package lfconfig

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("errors when fail to read file from path", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := Load(path.Join(tmpDir, "this-file-do-not-exist"))
		assert.Error(t, err)
	})

	t.Run("errors when fail to decode file", func(t *testing.T) {
		tmpDir := t.TempDir()
		fpath := path.Join(tmpDir, "config")
		err := os.WriteFile(fpath, []byte(`"not" "valid" "yaml"`), 0644)
		require.NoError(t, err)

		_, err = Load(fpath)
		assert.Error(t, err)
	})

	t.Run("errors when current context does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		fpath := path.Join(tmpDir, "config")
		err := os.WriteFile(fpath, []byte(`currentContext: context`), 0644)
		require.NoError(t, err)

		_, err = Load(fpath)
		assert.Error(t, err)
	})

	t.Run("load config successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		fpath := path.Join(tmpDir, "config")
		raw := `current-context: context
contexts:
  context:
    type: local
    dir: test-dir`

		err := os.WriteFile(fpath, []byte(raw), 0644)
		require.NoError(t, err)

		cfg, err := Load(fpath)
		require.NoError(t, err)

		assert.Equal(t, "context", cfg.CurrentContext)
		assert.Equal(t, map[string]configContext{"context": {Type: "local", Dir: "test-dir"}}, cfg.Contexts)
	})
}
