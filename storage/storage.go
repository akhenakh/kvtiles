package storage

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("tiles not found")

type TileStore interface {
	ReadTileData(ctx context.Context, z uint8, x uint64, y uint64) ([]byte, error)
	LoadMapInfos() (MapInfos, error)
}

// MapInfos used to store information about the map if any in DB
type MapInfos struct {
	CenterLat   float64
	CenterLng   float64
	MinZoom     int
	MaxZoom     int
	Attribution string
	Name        string
}
