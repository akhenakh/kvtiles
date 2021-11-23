package pogreb

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*Storage, func()) {
	logger := log.NewLogfmtLogger(os.Stdout)

	tmpDir, err := ioutil.TempDir(os.TempDir(), "kvtiles-test-")
	require.NoError(t, err)

	wstorage, wclose, err := NewStorage(tmpDir, logger)
	require.NoError(t, err)

	database, err := sql.Open("sqlite", "../../testdata/hawaii.mbtiles")
	require.NoError(t, err)

	err = wstorage.StoreMap(database, 21.315603, -157.858093, 11, "hawaii")
	require.NoError(t, err)

	err = wclose()
	require.NoError(t, err)

	storage, close, err := NewStorage(tmpDir, logger)
	require.NoError(t, err)

	return storage, func() {
		close()
		os.Remove(tmpDir)
	}
}
