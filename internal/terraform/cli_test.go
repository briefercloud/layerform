package terraform

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/ergomake/layerform/mocks/internal_/cmdexec"
)

func TestTerraformCLI_Init(t *testing.T) {
	tmpDir := t.TempDir()

	exec := mocks.NewCommandExecutor(t)

	exec.EXPECT().SetDir(tmpDir).Return()

	expectedErr := errors.New("testing error")
	exec.EXPECT().Run("terraform", "init").Return(0, expectedErr)

	cli := NewCLI(exec)
	err := cli.Init(tmpDir)

	assert.Equal(t, expectedErr, err)
}

func TestTerraformCLI_Apply(t *testing.T) {
	t.Run("writes state to disk and calls terraform apply", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		nextStateBytes := []byte(`{ "resources": [] }`)
		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().
			Run("terraform", "apply").
			RunAndReturn(func(_ string, _ ...string) (int, error) {
				err := os.WriteFile(path.Join(tmpDir, "terraform.tfstate"), nextStateBytes, 0644)

				return 0, err
			})

		cli := NewCLI(exec)
		nextState, err := cli.Apply(tmpDir, &State{Bytes: []byte("curr state")})

		require.NoError(t, err)
		require.NotNil(t, nextState)

		assert.Equal(t, string(nextStateBytes), string(nextState.Bytes))
	})

	t.Run("ignores nil state", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().
			Run("terraform", "apply").
			RunAndReturn(func(_ string, _ ...string) (int, error) {
				assert.NoFileExists(t, path.Join(tmpDir, "terraform.tfstate"))

				err := os.WriteFile(path.Join(tmpDir, "terraform.tfstate"), []byte("next state"), 0644)

				return 0, err
			})

		cli := NewCLI(exec)
		cli.Apply(tmpDir, nil)
	})

	t.Run("errors when fail to write state", func(t *testing.T) {
		tmpDir := t.TempDir()
		// remove tmp dir so apply fails to write state there
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)

		exec := mocks.NewCommandExecutor(t)

		cli := NewCLI(exec)
		_, err = cli.Apply(tmpDir, &State{Bytes: []byte("curr state")})
		assert.Error(t, err)
	})

	t.Run("errors when terraform apply errors", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		expectedErr := errors.New("apply error")
		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().Run("terraform", "apply").Return(0, expectedErr)

		cli := NewCLI(exec)
		_, err := cli.Apply(tmpDir, nil)
		assert.ErrorIs(t, expectedErr, err)
	})

	t.Run("errors when cant read new state", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		exec.EXPECT().SetDir(tmpDir).Return()
		// mock does not write state to disk to force failure
		exec.EXPECT().Run("terraform", "apply").Return(0, nil)

		cli := NewCLI(exec)
		_, err := cli.Apply(tmpDir, nil)
		assert.Error(t, err)
	})
}

func TestTerraformCLI_Destroy(t *testing.T) {
	t.Run("writes state to disk and calls terraform destroy", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		nextStateBytes := []byte(`{ "resources": [] }`)
		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().
			Run("terraform", "destroy").
			RunAndReturn(func(_ string, _ ...string) (int, error) {
				err := os.WriteFile(path.Join(tmpDir, "terraform.tfstate"), nextStateBytes, 0644)

				return 0, err
			})

		cli := NewCLI(exec)
		nextState, err := cli.Destroy(tmpDir, &State{Bytes: []byte("curr state")})

		require.NoError(t, err)
		require.NotNil(t, nextStateBytes)

		assert.FileExists(t, path.Join(tmpDir, "terraform.tfstate"))
		assert.Equal(t, string(nextStateBytes), string(nextState.Bytes))
	})

	t.Run("calls terraform destroy with target", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().
			Run("terraform", "destroy", "-target", "resource1", "-target", "resource2").
			RunAndReturn(func(_ string, _ ...string) (int, error) {
				err := os.WriteFile(path.Join(tmpDir, "terraform.tfstate"), []byte(`{ "resources": [] }`), 0644)
				return 0, err
			})

		cli := NewCLI(exec)
		_, err := cli.Destroy(tmpDir, &State{Bytes: []byte("curr state")}, "resource1", "resource2")

		require.NoError(t, err)
	})

	t.Run("ignores nil state", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().
			Run("terraform", "destroy").
			RunAndReturn(func(_ string, _ ...string) (int, error) {
				assert.NoFileExists(t, path.Join(tmpDir, "terraform.tfstate"))

				err := os.WriteFile(path.Join(tmpDir, "terraform.tfstate"), []byte("next state"), 0644)

				return 0, err
			})

		cli := NewCLI(exec)
		cli.Destroy(tmpDir, nil)
	})

	t.Run("errors when fail to write state", func(t *testing.T) {
		tmpDir := t.TempDir()
		// remove tmp dir so destroy fails to write state there
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)

		exec := mocks.NewCommandExecutor(t)

		cli := NewCLI(exec)
		_, err = cli.Destroy(tmpDir, &State{Bytes: []byte("curr state")})
		assert.Error(t, err)
	})

	t.Run("errors when terraform destroy errors", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		expectedErr := errors.New("destroy error")
		exec.EXPECT().SetDir(tmpDir).Return()
		exec.EXPECT().Run("terraform", "destroy").Return(0, expectedErr)

		cli := NewCLI(exec)
		_, err := cli.Destroy(tmpDir, nil)
		assert.ErrorIs(t, expectedErr, err)
	})

	t.Run("errors when cant read new state", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := mocks.NewCommandExecutor(t)

		exec.EXPECT().SetDir(tmpDir).Return()
		// mock does not write state to disk to force failure
		exec.EXPECT().Run("terraform", "destroy").Return(0, nil)

		cli := NewCLI(exec)
		_, err := cli.Destroy(tmpDir, nil)
		assert.Error(t, err)
	})

}
