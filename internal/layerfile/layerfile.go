package layerfile

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
)

var (
	ErrInvalidDefinitionName  = errors.New("invalid layer definition name")
	ErrDependencyDoesNotExist = errors.New("dependency does not exist")
)

var alphanumericRegex = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9_-]*[A-Za-z0-9]$")

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

func (lf *layerfile) ToLayers() ([]*data.LayerDefinition, error) {
	if err := lf.validateLayersDependencies(); err != nil {
		return nil, errors.Wrap(err, "fail to validate layers dependencies")
	}

	dir := path.Dir(lf.sourceFilepath)

	dataLayers := make([]*data.LayerDefinition, len(lf.Layers))
	for i, l := range lf.Layers {
		if !alphanumericRegex.MatchString(l.Name) {
			return nil, errors.Wrap(ErrInvalidDefinitionName, l.Name)
		}

		files := []data.LayerDefinitionFile{}
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

				files = append(files, data.LayerDefinitionFile{
					Path:    rel,
					Content: content,
				})
			}
		}

		layer := &data.LayerDefinition{
			Name:         l.Name,
			Files:        files,
			Dependencies: l.Dependencies,
		}
		sha, err := data.LayerDefinitionSHA(layer)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to compute sha1 of layer %s", l.Name)
		}
		layer.SHA = sha

		dataLayers[i] = layer
	}

	return dataLayers, nil
}

func (lf *layerfile) validateLayersDependencies() error {
	names := make(map[string]struct{})

	for _, l := range lf.Layers {
		names[l.Name] = struct{}{}
	}

	for _, l := range lf.Layers {
		for _, d := range l.Dependencies {
			if _, ok := names[d]; !ok {
				return errors.Wrap(ErrDependencyDoesNotExist, d)
			}
		}
	}

	return nil
}
