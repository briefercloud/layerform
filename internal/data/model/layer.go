package model

import (
	"encoding/json"
	"io"
)

type Layer struct {
	Name         string      `json:"name"`
	Files        []LayerFile `json:"files"`
	Dependencies []string    `json:"dependencies"`
}

func (c *Layer) FromJSON(data io.Reader) error {
	de := json.NewDecoder(data)
	return de.Decode(c)
}

func (c *Layer) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

type LayerFile struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}
