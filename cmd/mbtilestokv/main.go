// +build cgo

package main

import (
	"database/sql"
	"os"
	"path"

	log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/namsral/flag"
	_ "modernc.org/sqlite"

	"github.com/akhenakh/kvtiles/loglevel"
	"github.com/akhenakh/kvtiles/storage"
	bstorage "github.com/akhenakh/kvtiles/storage/bbolt"
	pstorage "github.com/akhenakh/kvtiles/storage/pogreb"
)

const appName = "mbtilestokv"

var (
	version  = "no version from LDFLAGS"
	logLevel = flag.String("logLevel", "INFO", "DEBUG|INFO|WARN|ERROR")

	centerLat   = flag.Float64("centerLat", 48.8, "Latitude center used for the debug map")
	centerLng   = flag.Float64("centerLng", 2.2, "Longitude center used for the debug map")
	maxZoom     = flag.Int("maxZoom", 9, "max zoom level")
	storageType = flag.String("storageType", "pogreb", "pogreb|bbolt")

	tilesPath = flag.String("tilesPath", "./hawaii.mbtiles", "mbtiles file path")
	dbPath    = flag.String("dbPath", "./map.db", "db path out")
)

func main() {
	flag.Parse()

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "caller", log.Caller(5), "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "app", appName)
	logger = loglevel.NewLevelFilterFromString(logger, *logLevel)

	level.Info(logger).Log("msg", "starting converting tiles", "version", version)

	database, err := sql.Open("sqlite", *tilesPath)
	if err != nil {
		level.Error(logger).Log("msg", "can't read mbtiles sqlite", "error", err)
		os.Exit(2)
	}
	defer database.Close()

	var storage storage.TileStore

	switch *storageType {
	case "bbolt":
		s, clean, err := bstorage.NewStorage(*dbPath, logger)
		if err != nil {
			level.Error(logger).Log("msg", "can't open storage for writing", "error", err)
			os.Exit(2)
		}
		storage = s
		defer clean()

	case "pogreb":
		s, clean, err := pstorage.NewStorage(*dbPath, logger)
		if err != nil {
			level.Error(logger).Log("msg", "can't open storage for writing", "error", err)
			os.Exit(2)
		}
		storage = s
		defer clean()
	}

	err = storage.StoreMap(database, *centerLat, *centerLng, *maxZoom, path.Base(*tilesPath))
	if err != nil {
		level.Error(logger).Log("msg", "can't store tiles in db", "error", err)
		os.Exit(2)
	}
}
