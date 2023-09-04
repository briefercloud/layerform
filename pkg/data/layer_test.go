package data

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayerDeserializeFromJSON(t *testing.T) {
	l := Layer{}

	err := l.FromJSON(bytes.NewReader([]byte(layerData)))
	require.NoError(t, err)

	assert.Equal(t, "base", l.Name)
	assert.Len(t, l.Files, 1)
	assert.Equal(t, "./base/main.tf", l.Files[0].Path)
	assert.Equal(t, []byte(strContent), l.Files[0].Content)
}

func TestLayerSerializesToJSON(t *testing.T) {
	l := Layer{
		Name: "base",
		Files: []LayerFile{
			{Path: "./base/main.tf", Content: []byte(strContent)},
		},
	}

	d, err := l.ToJSON()
	require.NoError(t, err)

	ld := make(map[string]interface{}, 0)
	err = json.Unmarshal(d, &ld)
	require.NoError(t, err)

	assert.Equal(t, "base", ld["name"])

	f := ld["files"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "./base/main.tf", f["path"])
	assert.Equal(t, base64Content, f["content"])
}

var strContent = `locals {
  name = "value"
}`

var base64Content = base64.StdEncoding.EncodeToString([]byte(strContent))

var layerData = `{
  "name": "base",
  "files": [
    { "path": "./base/main.tf", "content": "` + base64Content + `" }
  ]
}`
