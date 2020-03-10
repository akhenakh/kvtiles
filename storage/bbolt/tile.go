package bbolt

import (
	"errors"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/akhenakh/kvtiles/storage"
)

// ReadTileData returns []bytes from a tile
func (s *Storage) ReadTileData(z uint8, x uint64, y uint64) ([]byte, error) {
	var v []byte
	err := s.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(storage.MapKey())

		k := []byte(fmt.Sprintf("%c%d/%d/%d", storage.TilesURLPrefix, z, x, y))
		v = b.Get(k)
		if v == nil {
			return nil
		}

		tk := []byte{storage.TilesPrefix}
		tk = append(tk, v...)
		v = b.Get(tk)
		if v == nil {
			return errors.New("can't find blob at existing entry")
		}
		return nil
	})

	return v, err
}
