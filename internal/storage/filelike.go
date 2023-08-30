package storage

import (
	"context"
)

type FileLike interface {
	Path(ctx context.Context) (string, error)
	Load(ctx context.Context, v any) error
	Save(ctx context.Context, v any) error
}
