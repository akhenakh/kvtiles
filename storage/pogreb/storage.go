package pogreb

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/akhenakh/kvtiles/storage"
	"github.com/akrylysov/pogreb"
	"github.com/cespare/xxhash"
	"github.com/fxamacker/cbor/v2"
	"github.com/go-kit/kit/log/level"
	log "github.com/go-kit/log"
)

// Storage cold storage.
type Storage struct {
	*pogreb.DB
	logger log.Logger
	TMS    bool
}

// NewStorage returns a cold storage using pogreb.
func NewStorage(path string, logger log.Logger) (*Storage, func() error, error) {
	// Creating DB
	db, err := pogreb.Open(path, nil)
	if err != nil {
		return nil, nil, err
	}

	return &Storage{
		DB:     db,
		logger: logger,
	}, db.Close, nil
}

// LoadMapInfos loads map infos from the DB if any.
func (s *Storage) LoadMapInfos() (*storage.MapInfos, bool, error) {
	var mapInfos *storage.MapInfos
	value, err := s.DB.Get(storage.MapKey())
	if err != nil {
		return nil, false, err
	}
	if value == nil {
		return nil, false, err
	}
	mapInfos = &storage.MapInfos{}
	dec := cbor.NewDecoder(bytes.NewReader(value))
	if err := dec.Decode(mapInfos); err != nil {
		return nil, false, err
	}

	if mapInfos == nil {
		return nil, false, nil
	}

	return mapInfos, true, nil
}

func (s *Storage) storeOldSchema(database *sql.DB, maxZoom int) error {
	rows, err := database.Query(
		"SELECT tile_id, zoom_level, tile_column, tile_row FROM map where zoom_level <= ?",
		maxZoom,
	)
	if err != nil {
		return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var zoom, column, row int
		var tileID string
		rows.Scan(&tileID, &zoom, &column, &row)

		urlkey := fmt.Sprintf("%c%d/%d/%d", storage.TilesURLPrefix, zoom, column, row)
		level.Debug(s.logger).Log("msg", "storing url", "url", urlkey, "tileID", tileID)

		if err = s.DB.Put([]byte(urlkey), []byte(tileID)); err != nil {
			return err
		}
	}

	trows, err := database.Query(
		"SELECT images.tile_data, images.tile_id from images JOIN  map ON images.tile_id = map.tile_id where zoom_level <= ?;",
		maxZoom)
	if err != nil {
		return err
	}

	if trows.Err() != nil {
		return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
	}
	defer trows.Close()

	for trows.Next() {
		var tileData []byte
		var tileID string

		for trows.Next() {
			trows.Scan(&tileData, &tileID)
			key := fmt.Sprintf("%s", tileID)
			if err = s.DB.Put([]byte(key), tileData); err != nil {
				return err
			}
			level.Debug(s.logger).Log("msg", "storing tileID", "tileID", tileID, "data_size", len(tileData))
		}
	}

	return nil
}

func (s *Storage) storeMapUtil(database *sql.DB, maxZoom int) error {
	rows, err := database.Query(
		"SELECT tile_data, zoom_level, tile_column, tile_row FROM tiles where zoom_level <= ?",
		maxZoom,
	)
	if err != nil {
		return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var zoom, column, row int
		var tileData []byte
		rows.Scan(&tileData, &zoom, &column, &row)

		// TODO: watchout for collisions
		thash := xxhash.Sum64(tileData)
		khash := make([]byte, 8)
		binary.BigEndian.PutUint64(khash, thash)

		key := fmt.Sprintf("%c%d/%d/%d", storage.TilesURLPrefix, zoom, column, row)
		if err = s.DB.Put([]byte(key), khash); err != nil {
			return err
		}

		if err = s.DB.Put(khash, tileData); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) StoreMap(database *sql.DB, centerLat, centerLng float64, maxZoom int, region string) error {
	oldSchema := false
	// find if we are using the old schema
	row := database.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='tiles';",
	)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			oldSchema = true
		} else {
			return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
		}
	}

	if !oldSchema {
		if err := s.storeMapUtil(database, maxZoom); err != nil {
			return fmt.Errorf("can't store map using mbutil format: %w", err)
		}
		level.Debug(s.logger).Log("msg", "using mbutil format")
	} else {
		if err := s.storeOldSchema(database, maxZoom); err != nil {
			return fmt.Errorf("can't store map using old schema format: %w", err)
		}
		level.Debug(s.logger).Log("msg", "using old format")
	}

	infos := &storage.MapInfos{
		CenterLat: centerLat,
		CenterLng: centerLng,
		MaxZoom:   maxZoom,
		Region:    region,
		IndexTime: time.Now(),
		TMS:       oldSchema,
	}

	infoBytes, err := cbor.Marshal(infos)
	if err != nil {
		return fmt.Errorf("failed encoding MapInfos: %w", err)
	}

	return s.DB.Put(storage.MapKey(), infoBytes)
}
