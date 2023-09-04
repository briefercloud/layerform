package data

import (
	"crypto/sha1"
	"sort"
)

type LayerDefinition struct {
	SHA          []byte           `json:"sha"`
	Name         string           `json:"name"`
	Files        []LayerDefinitionFile `json:"files"`
	Dependencies []string         `json:"dependencies"`
}

type LayerDefinitionFile struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

func LayerDefinitionSHA(l *LayerDefinition) ([]byte, error) {
	hasher := sha1.New()
	for _, f := range l.Files {
		_, err := hasher.Write([]byte("path:" + f.Path + "\n"))
		if err != nil {
			return nil, err
		}

		_, err = hasher.Write([]byte("content:"))
		if err != nil {
			return nil, err
		}

		_, err = hasher.Write(f.Content)
		if err != nil {
			return nil, err
		}

		_, err = hasher.Write([]byte("\n"))
		if err != nil {
			return nil, err
		}
	}

	deps := make([]string, len(l.Dependencies))
	copy(deps, l.Dependencies)
	sort.Strings(deps)

	_, err := hasher.Write([]byte("deps:"))
	if err != nil {
		return nil, err
	}

	for _, d := range deps {
		_, err := hasher.Write([]byte(d))
		if err != nil {
			return nil, err
		}
	}

	return hasher.Sum(nil), nil
}
