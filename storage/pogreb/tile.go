package pogreb

import (
	"errors"
	"fmt"

	"github.com/akhenakh/kvtiles/storage"
)

// ReadTileData returns []bytes from a tile
func (s *Storage) ReadTileData(z uint8, x uint64, y uint64) ([]byte, error) {
	k := []byte(fmt.Sprintf("%c%d/%d/%d", storage.TilesURLPrefix, z, x, y))
	v, err := s.DB.Get(k)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	v, err = s.DB.Get(v)
	if v == nil {
		return nil, errors.New("can't find blob at existing entry")
	}

	return v, err
}
