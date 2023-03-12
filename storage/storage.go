package storage

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("tiles not found")

type TileStore interface {
	ReadTileData(ctx context.Context, z uint8, x uint64, y uint64) ([]byte, error)
}
