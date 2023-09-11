package layerfile

import (
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	t.Run("parses layers from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `{
  "layers": [
    {
      "name": "layer1",
      "files": ["main.tf"],
      "dependencies": []
    }
  ]
}`

		filePath := path.Join(tmpDir, "layerform.json")

		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		lf, err := FromFile(filePath)
		require.NoError(t, err)

		assert.NotNil(t, lf)
		assert.Equal(t, 1, len(lf.Layers))
		assert.Equal(t, "layer1", lf.Layers[0].Name)
		assert.Equal(t, 1, len(lf.Layers[0].Files))
		assert.Equal(t, "main.tf", lf.Layers[0].Files[0])
		assert.Equal(t, 0, len(lf.Layers[0].Dependencies))
	})

	t.Run("fails to read file", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := path.Join(tmpDir, "layerform.json")

		// did not actually create the file

		_, err := FromFile(filePath)
		assert.Error(t, err)
	})
}

func TestToLayers(t *testing.T) {
	tmpDir := t.TempDir()

	sourceFilePath := path.Join(tmpDir, "layerform.json")
	err := os.WriteFile(sourceFilePath, []byte("{}"), 0644)
	require.NoError(t, err)

	mainTfContent := []byte("main.tf content")
	err = os.WriteFile(path.Join(tmpDir, "main.tf"), mainTfContent, 0644)
	require.NoError(t, err)

	lf := &layerfile{
		sourceFilepath: sourceFilePath,
		Layers: []layerfileLayer{
			{
				Name:         "layer1",
				Files:        []string{"main.tf"},
				Dependencies: make([]string, 0),
			},
		},
	}

	modelLayers, err := lf.ToLayers()
	require.NoError(t, err)

	assert.NotNil(t, modelLayers)
	assert.Equal(t, 1, len(modelLayers))
	assert.Equal(t, "layer1", modelLayers[0].Name)
	assert.Equal(t, 1, len(modelLayers[0].Files))
	assert.Equal(t, "main.tf", modelLayers[0].Files[0].Path)
	assert.Equal(t, mainTfContent, modelLayers[0].Files[0].Content)
	assert.Equal(t, 0, len(modelLayers[0].Dependencies))
}

func TestToLayers_ValidateNameOfLayerDefinitions(t *testing.T) {
	tests := []struct {
		name string
		lf   layerfile
		err  error
	}{
		{
			name: "Name has spaces",
			lf: layerfile{
				Layers: []layerfileLayer{
					{
						Name: "invalid name for a layer definition",
					},
				},
			},
			err: errors.New("invalid name: invalid name for a layer definition"),
		},
		{
			name: "Name has special character",
			lf: layerfile{
				Layers: []layerfileLayer{
					{
						Name: "invalid!",
					},
				},
			},
			err: errors.New("invalid name: invalid!"),
		},
		{
			name: "Valid name",
			lf: layerfile{
				Layers: []layerfileLayer{
					{
						Name: "validname",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.lf.ToLayers()
			if err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				assert.NoError(t, tt.err)
			}
		})
	}
}
