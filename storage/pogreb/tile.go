package pogreb

import (
	"fmt"
	"sync"

	"github.com/akhenakh/kvtiles/storage"
	"github.com/go-kit/kit/log/level"
)

var doOnce sync.Once

// ReadTileData returns []bytes from a tile
func (s *Storage) ReadTileData(z uint8, x uint64, y uint64) ([]byte, error) {
	doOnce.Do(func() {
		mi, _, _ := s.LoadMapInfos()
		s.TMS = mi.TMS
	})

	if s.TMS {
		y = uint64(1<<uint(z) - y - 1)
	}
	k := []byte(fmt.Sprintf("%c%d/%d/%d", storage.TilesURLPrefix, z, x, y))

	level.Debug(s.logger).Log("msg", "read tile", "url_key", string(k))
	v, err := s.DB.Get(k)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	nv, err := s.DB.Get(v)
	if nv == nil {
		return nil, fmt.Errorf("can't find blob at existing entry %s %s", string(k), string(v))
	}

	return nv, err
}
