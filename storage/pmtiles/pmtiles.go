package pmtiles

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path"

	"github.com/akhenakh/kvtiles/storage"
	log "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gocloud.dev/blob"

	"github.com/protomaps/go-pmtiles/pmtiles"
)

const HeaderV3Len = 127

type HeaderV3 struct {
	SpecVersion         uint8
	RootOffset          uint64
	RootLength          uint64
	MetadataOffset      uint64
	MetadataLength      uint64
	LeafDirectoryOffset uint64
	LeafDirectoryLength uint64
	TileDataOffset      uint64
	TileDataLength      uint64
	AddressedTilesCount uint64
	TileEntriesCount    uint64
	TileContentsCount   uint64
	Clustered           bool
	InternalCompression Compression
	TileCompression     Compression
	TileType            TileType
	MinZoom             uint8
	MaxZoom             uint8
	MinLonE7            int32
	MinLatE7            int32
	MaxLonE7            int32
	MaxLatE7            int32
	CenterZoom          uint8
	CenterLonE7         int32
	CenterLatE7         int32
}

type TileType uint8

const (
	UnknownTileType TileType = 0
	Mvt                      = 1
	Png                      = 2
	Jpeg                     = 3
	Webp                     = 4
)

type Compression uint8

const (
	UnknownCompression Compression = 0
	NoCompression                  = 1
	Gzip                           = 2
	Brotli                         = 3
	Zstd                           = 4
)

type EntryV3 struct {
	TileId    uint64
	Offset    uint64
	Length    uint32
	RunLength uint32
}

type Storage struct {
	bucket *blob.Bucket
	header HeaderV3
	file   string
}

func NewStorage(ctx context.Context, logger log.Logger, url string) (func(), *Storage, error) {
	dir := path.Dir(url)
	bucket, err := blob.OpenBucket(ctx, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open bucket for %s, %w", url, err)
	}

	clean := func() { bucket.Close() }

	file := path.Base(url)
	r, err := bucket.NewRangeReader(ctx, file, 0, 16384, nil)
	if err != nil {
		return clean, nil, fmt.Errorf("failed to create range reader for %s, %w", file, err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return clean, nil, fmt.Errorf("failed to read %s, %w", file, err)
	}
	r.Close()

	header, err := deserialize_header(b[0:HeaderV3Len])
	if err != nil {
		return clean, nil, fmt.Errorf("failed to read %s, %w", file, err)
	}

	level.Debug(logger).Log("msg", "storage opened", "file", file, "tile_type", header.TileType)

	s := &Storage{
		bucket: bucket,
		header: header,
		file:   file,
	}

	return clean, s, nil
}

func (s *Storage) ReadTileData(ctx context.Context, z uint8, x uint64, y uint64) ([]byte, error) {
	tile_id := pmtiles.ZxyToId(uint8(z), uint32(x), uint32(y))

	dir_offset := s.header.RootOffset
	dir_length := s.header.RootLength

	for depth := 0; depth <= 3; depth++ {
		r, err := s.bucket.NewRangeReader(ctx, s.file, int64(dir_offset), int64(dir_length), nil)
		if err != nil {
			return nil, fmt.Errorf("range reader error: %w", err)
		}
		defer r.Close()

		b, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("can't read bucket: %w", err)
		}

		directory := deserialize_entries(bytes.NewBuffer(b))
		entry, ok := find_tile(directory, tile_id)

		if ok {
			if entry.RunLength > 0 {
				tile_r, err := s.bucket.NewRangeReader(
					ctx,
					s.file,
					int64(s.header.TileDataOffset+entry.Offset),
					int64(entry.Length),
					nil,
				)
				if err != nil {
					return nil, fmt.Errorf("can't read bucket: %w", err)
				}
				defer tile_r.Close()

				tile_b, err := io.ReadAll(tile_r)
				if err != nil {
					return nil, fmt.Errorf("can't read bucket: %w", err)
				}

				return tile_b, nil
			} else {
				dir_offset = s.header.LeafDirectoryOffset + entry.Offset
				dir_length = uint64(entry.Length)
			}
		} else {
			return nil, storage.ErrNotFound
		}
	}

	return nil, storage.ErrNotFound
}

func deserialize_header(d []byte) (HeaderV3, error) {
	h := HeaderV3{}
	magic_number := d[0:7]
	if string(magic_number) != "PMTiles" {
		return h, fmt.Errorf("Magic number not detected. Are you sure this is a PMTiles archive?")
	}

	spec_version := d[7]
	if spec_version > uint8(3) {
		return h, fmt.Errorf("Archive is spec version %d, but this program only supports version 3: upgrade your pmtiles program.", spec_version)
	}

	h.SpecVersion = spec_version
	h.RootOffset = binary.LittleEndian.Uint64(d[8 : 8+8])
	h.RootLength = binary.LittleEndian.Uint64(d[16 : 16+8])
	h.MetadataOffset = binary.LittleEndian.Uint64(d[24 : 24+8])
	h.MetadataLength = binary.LittleEndian.Uint64(d[32 : 32+8])
	h.LeafDirectoryOffset = binary.LittleEndian.Uint64(d[40 : 40+8])
	h.LeafDirectoryLength = binary.LittleEndian.Uint64(d[48 : 48+8])
	h.TileDataOffset = binary.LittleEndian.Uint64(d[56 : 56+8])
	h.TileDataLength = binary.LittleEndian.Uint64(d[64 : 64+8])
	h.AddressedTilesCount = binary.LittleEndian.Uint64(d[72 : 72+8])
	h.TileEntriesCount = binary.LittleEndian.Uint64(d[80 : 80+8])
	h.TileContentsCount = binary.LittleEndian.Uint64(d[88 : 88+8])
	h.Clustered = (d[96] == 0x1)
	h.InternalCompression = Compression(d[97])
	h.TileCompression = Compression(d[98])
	h.TileType = TileType(d[99])
	h.MinZoom = d[100]
	h.MaxZoom = d[101]
	h.MinLonE7 = int32(binary.LittleEndian.Uint32(d[102 : 102+4]))
	h.MinLatE7 = int32(binary.LittleEndian.Uint32(d[106 : 106+4]))
	h.MaxLonE7 = int32(binary.LittleEndian.Uint32(d[110 : 110+4]))
	h.MaxLatE7 = int32(binary.LittleEndian.Uint32(d[114 : 114+4]))
	h.CenterZoom = d[118]
	h.CenterLonE7 = int32(binary.LittleEndian.Uint32(d[119 : 119+4]))
	h.CenterLatE7 = int32(binary.LittleEndian.Uint32(d[123 : 123+4]))

	return h, nil
}

func deserialize_entries(data *bytes.Buffer) []EntryV3 {
	entries := make([]EntryV3, 0)

	reader, _ := gzip.NewReader(data)
	byte_reader := bufio.NewReader(reader)

	num_entries, _ := binary.ReadUvarint(byte_reader)

	last_id := uint64(0)
	for i := uint64(0); i < num_entries; i++ {
		tmp, _ := binary.ReadUvarint(byte_reader)
		entries = append(entries, EntryV3{last_id + tmp, 0, 0, 0})
		last_id = last_id + tmp
	}

	for i := uint64(0); i < num_entries; i++ {
		run_length, _ := binary.ReadUvarint(byte_reader)
		entries[i].RunLength = uint32(run_length)
	}

	for i := uint64(0); i < num_entries; i++ {
		length, _ := binary.ReadUvarint(byte_reader)
		entries[i].Length = uint32(length)
	}

	for i := uint64(0); i < num_entries; i++ {
		tmp, _ := binary.ReadUvarint(byte_reader)
		if i > 0 && tmp == 0 {
			entries[i].Offset = entries[i-1].Offset + uint64(entries[i-1].Length)
		} else {
			entries[i].Offset = tmp - 1
		}
	}

	return entries
}

func find_tile(entries []EntryV3, tileId uint64) (EntryV3, bool) {
	m := 0
	n := len(entries) - 1
	for m <= n {
		k := (n + m) >> 1
		cmp := int64(tileId) - int64(entries[k].TileId)
		if cmp > 0 {
			m = k + 1
		} else if cmp < 0 {
			n = k - 1
		} else {
			return entries[k], true
		}
	}

	// at this point, m > n
	if n >= 0 {
		if entries[n].RunLength == 0 {
			return entries[n], true
		}
		if tileId-entries[n].TileId < uint64(entries[n].RunLength) {
			return entries[n], true
		}
	}
	return EntryV3{}, false
}
