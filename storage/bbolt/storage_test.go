package bbolt

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*Storage, func()) {
	logger := log.NewLogfmtLogger(os.Stdout)

	tmpFile, err := ioutil.TempFile(os.TempDir(), "kvtiles-test-")
	require.NoError(t, err)
	wstorage, wclose, err := NewStorage(tmpFile.Name(), logger)

	database, err := sql.Open("sqlite3", "../../testdata/hawaii.mbtiles")
	require.NoError(t, err)

	err = wstorage.StoreMap(database, 21.315603, -157.858093, 11, "hawaii")
	require.NoError(t, err)

	err = wclose()
	require.NoError(t, err)

	// RO storage
	storage, close, err := NewROStorage(tmpFile.Name(), logger)
	require.NoError(t, err)

	return storage, func() {
		close()
		os.Remove(tmpFile.Name())
	}
}
