//go:build cgo
// +build cgo

package pogreb

import (
	"encoding/base64"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestStorage_ReadTileData(t *testing.T) {
	s, clean := setup(t)
	defer clean()

	tests := []struct {
		name    string
		z       uint8
		x       uint64
		y       uint64
		want    string
		wantErr bool
	}{
		{
			"regular read",
			11, 124, 900,
			"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",
			false,
		},
		{
			"non existing read",
			12, 124, 900,
			"",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := s.ReadTileData(tt.z, tt.x, tt.y)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadTileData() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.want == "" {
				require.Empty(t, got)

				return
			}

			decoded, err := base64.StdEncoding.DecodeString(tt.want)
			if err != nil {
				t.Errorf("error decoding b64 %s", err)
			}

			if !cmp.Equal(got, decoded) {
				t.Errorf("ReadTileData() got = %v, want %v", got, decoded)
			}
		})
	}
}
