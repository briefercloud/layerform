package storage

import (
	"context"
)

type FileLike interface {
	Load(ctx context.Context, v any) error
	Save(ctx context.Context, v any) error
}
