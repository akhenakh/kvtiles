package storage

import (
	"database/sql"
	"time"
)

const (
	mapKey byte = 'm'
	// reserved T & t for tiles
	TilesURLPrefix byte = 't'
	TilesPrefix    byte = 'T'
)

type TileStore interface {
	ReadTileData(z uint8, x uint64, y uint64) ([]byte, error)
	StoreMap(database *sql.DB, centerLat, centerLng float64, maxZoom int, region string) error
}

// MapInfos used to store information about the map if any in DB
type MapInfos struct {
	CenterLat, CenterLng float64
	MaxZoom              int
	Region               string
	IndexTime            time.Time
}

// MapKey returns the key for the map entry
func MapKey() []byte {
	return []byte{mapKey}
}
