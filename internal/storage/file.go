package storage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

type fileStorage struct {
	fpath string
}

var _ FileLike = &fileStorage{}

func NewFileStorage(fpath string) *fileStorage {
	return &fileStorage{fpath}
}

func (fls *fileStorage) Load(ctx context.Context, v any) error {
	hclog.FromContext(ctx).Debug("Reading layers file", "path", fls.fpath)

	raw, err := os.ReadFile(fls.fpath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return errors.Wrapf(err, "fail to read %s", fls.fpath)
	}

	err = json.Unmarshal(raw, &v)
	return errors.Wrapf(err, "fail to parse layers out of %s", fls.fpath)

}

func (fls *fileStorage) Save(ctx context.Context, v any) error {
	hclog.FromContext(ctx).Debug("Writting layers to file", "path", fls.fpath)

	data, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "fail to marshal filelayers")
	}

	err = os.WriteFile(fls.fpath, data, 0644)
	return errors.Wrap(err, "fail to write file")
}
