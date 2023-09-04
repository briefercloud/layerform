package data

import (
	"crypto/sha1"
	"encoding/json"
	"io"
	"sort"

	"github.com/pkg/errors"
)

type Layer struct {
	SHA          []byte      `json:"sha"`
	Name         string      `json:"name"`
	Files        []LayerFile `json:"files"`
	Dependencies []string    `json:"dependencies"`
}

func (l *Layer) FromJSON(data io.Reader) error {
	de := json.NewDecoder(data)
	return de.Decode(l)
}

func (l *Layer) ToJSON() ([]byte, error) {
	sha, err := LayerSHA(l)
	if err != nil {
		return nil, errors.Wrap(err, "fail to compute layer sha")
	}

	l.SHA = sha
	return json.Marshal(l)
}

type LayerFile struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

func LayerSHA(l *Layer) ([]byte, error) {
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
