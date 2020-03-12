package bbolt

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"

	"github.com/akhenakh/kvtiles/storage"
	"github.com/fxamacker/cbor"
	log "github.com/go-kit/kit/log"
	"go.etcd.io/bbolt"
)

// Storage cold storage
type Storage struct {
	*bbolt.DB
	logger log.Logger
}

// NewStorage returns a cold storage using bboltdb
func NewStorage(path string, logger log.Logger) (*Storage, func() error, error) {
	// Creating DB
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, nil, err
	}

	return &Storage{
		DB:     db,
		logger: logger,
	}, db.Close, nil
}

// NewROStorage returns a read only storage using bboltdb
func NewROStorage(path string, logger log.Logger) (*Storage, func() error, error) {
	// Creating DB
	db, err := bbolt.Open(path, 0600, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open DB for reading at %s: %w", path, err)
	}

	s := &Storage{
		DB:     db,
		logger: logger,
	}

	return s, db.Close, nil
}

// LoadMapInfos loads map infos from the DB if any
func (s *Storage) LoadMapInfos() (*storage.MapInfos, bool, error) {
	var mapInfos *storage.MapInfos
	err := s.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(storage.MapKey())
		if b == nil {
			return nil
		}
		value := b.Get(storage.MapKey())
		if value == nil {
			return nil
		}
		mapInfos = &storage.MapInfos{}
		dec := cbor.NewDecoder(bytes.NewReader(value))
		if err := dec.Decode(mapInfos); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}

	if mapInfos == nil {
		return nil, false, nil
	}

	return mapInfos, true, nil
}

func (s *Storage) StoreMap(database *sql.DB, centerLat, centerLng float64, maxZoom int, region string) error {
	rows, err := database.Query("SELECT * FROM map where zoom_level <= ?", maxZoom)
	if err != nil {
		return fmt.Errorf("can't read data from mbtiles sqlite: %w", err)
	}

	if err := s.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(storage.MapKey())
		if err != nil {
			return err
		}
		var zoom, column, row int
		var tileID, gridID, key string
		for rows.Next() {
			rows.Scan(&zoom, &column, &row, &tileID, &gridID)
			key = fmt.Sprintf("%c%d/%d/%d", storage.TilesURLPrefix, zoom, column, row)
			if err = b.Put([]byte(key), []byte(tileID)); err != nil {
				return err
			}
		}

		rows, err = database.Query("SELECT images.tile_data, images.tile_id from images JOIN  map ON images.tile_id = map.tile_id where zoom_level <= ?;", maxZoom)
		if err != nil {
			return err
		}

		var tileData []byte
		for rows.Next() {
			rows.Scan(&tileData, &tileID)
			key = fmt.Sprintf("%c%s", storage.TilesPrefix, tileID)
			if err = b.Put([]byte(key), tileData); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed writing to DB: %w", err)
	}

	infoBytes := new(bytes.Buffer)

	infos := &storage.MapInfos{
		CenterLat: centerLat,
		CenterLng: centerLng,
		MaxZoom:   maxZoom,
		Region:    region,
		IndexTime: time.Now(),
	}

	enc := cbor.NewEncoder(infoBytes, cbor.CanonicalEncOptions())
	if err := enc.Encode(infos); err != nil {
		return fmt.Errorf("failed encoding MapInfos: %w", err)
	}

	err = s.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(storage.MapKey())
		return b.Put(storage.MapKey(), infoBytes.Bytes())
	})
	if err != nil {
		return fmt.Errorf("failed writing MapInfos to DB: %w", err)
	}

	return nil
}
