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
	LoadMapInfos() (*MapInfos, bool, error)
	ReadTileData(z uint8, x uint64, y uint64) ([]byte, error)
	StoreMap(database *sql.DB, centerLat, centerLng float64, maxZoom int, region string) error
}

// MapInfos used to store information about the map if any in DB
type MapInfos struct {
	CenterLat float64   `cbor:"1,keyasint,omitempty"`
	CenterLng float64   `cbor:"2,keyasint,omitempty"`
	MaxZoom   int       `cbor:"3,keyasint,omitempty"`
	Region    string    `cbor:"4,keyasint,omitempty"`
	IndexTime time.Time `cbor:"5,keyasint,omitempty"`
	TMS       bool      `cbor:"6,keyasint,omitempty"`
}

// MapKey returns the key for the map entry
func MapKey() []byte {
	return []byte{mapKey}
}
