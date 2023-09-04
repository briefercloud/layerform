package layerfile

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
)

type layerfile struct {
	sourceFilepath string           `json:"-"`
	Layers         []layerfileLayer `json:"layers"`
}

type layerfileLayer struct {
	Name         string   `json:"name"`
	Files        []string `json:"files"`
	Dependencies []string `json:"dependencies"`
}

func FromFile(sourceFilepath string) (*layerfile, error) {
	bs, err := os.ReadFile(sourceFilepath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to read %s", sourceFilepath)
	}

	lf := &layerfile{sourceFilepath: sourceFilepath}
	err = json.Unmarshal(bs, lf)

	return lf, errors.Wrapf(err, "fail to decode %s into layerfile", lf)
}

func (lf *layerfile) ToLayers() ([]*data.Layer, error) {
	dir := path.Dir(lf.sourceFilepath)

	dataLayers := make([]*data.Layer, len(lf.Layers))
	for i, l := range lf.Layers {
		files := []data.LayerFile{}
		for _, f := range l.Files {
			matches, err := filepath.Glob(path.Join(dir, f))
			if err != nil {
				return nil, errors.Wrapf(err, "fail to apply glob pattern %s", f)
			}

			for _, fpath := range matches {
				content, err := os.ReadFile(fpath)
				if err != nil {
					return nil, errors.Wrapf(err, "could not read %s", fpath)
				}

				rel, err := filepath.Rel(dir, fpath)
				if err != nil {
					return nil, errors.Wrap(err, "fail to extract relative path")
				}

				files = append(files, data.LayerFile{
					Path:    rel,
					Content: content,
				})
			}
		}

		layer := &data.Layer{
			Name:         l.Name,
			Files:        files,
			Dependencies: l.Dependencies,
		}
		sha, err := data.LayerSHA(layer)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to compute sha1 of layer %s", l.Name)
		}
		layer.SHA = sha

		dataLayers[i] = layer
	}

	return dataLayers, nil
}
