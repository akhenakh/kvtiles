package server

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/akhenakh/kvtiles/storage"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
)

var templatesNames = []string{"protomap.style", "osm-liberty.style", "osm-liberty-gl.style", "planet.json", "index.html"}

// ServeHTTP serves the mbtiles for URL such as /tiles/11/618/722.pbf
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := log.With(s.logger, "component", "tile_server")
	vars := mux.Vars(req)

	z, _ := strconv.Atoi(vars["z"])
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])

	q := req.URL.Query()
	if s.tilesKey != "" {
		k := q.Get("key")
		if k != s.tilesKey {
			level.Debug(logger).Log("err", "unauthorized tile request")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}
	}

	data, err := s.tileStorage.ReadTileData(req.Context(), uint8(z), uint64(x), uint64(y))
	if err != nil {
		if err == storage.ErrNotFound {
			level.Debug(logger).Log(
				"err", "tile not found",
				"x", x,
				"z", z,
				"y", y,
			)

			http.NotFound(w, req)

			return
		}
		level.Debug(logger).Log(
			"err", err.Error(),
			"x", x,
			"z", z,
			"y", y,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if len(data) == 0 {
		http.NotFound(w, req)

		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Set("Content-Encoding", "gzip")
	_, _ = w.Write(data)
}

// TilesHandler serves the mbtiles at /tiles/11/618/722.pbf
func (s *Server) TilesHandler(w http.ResponseWriter, req *http.Request) {
	s.ServeHTTP(w, req)
}

// StaticHandler serves templates and other static files
func (s *Server) StaticHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// serve file normally
	if !isTpl(path) {
		s.fileHandler.ServeHTTP(w, req)

		return
	}

	// check for key if needed
	q := req.URL.Query()
	if s.tilesKey != "" {
		k := q.Get("key")
		if k != s.tilesKey {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}
	}

	// Templates variables
	proto := "http"
	if req.Header.Get("X-Forwarded-Proto") == "https" {
		proto = "https"
	}

	p := map[string]interface{}{
		"TilesBaseURL": fmt.Sprintf("%s://%s", proto, req.Host),
		"MaxZoom":      s.infos.MaxZoom,
		"CenterLat":    s.infos.CenterLat,
		"CenterLng":    s.infos.CenterLng,
		"TilesKey":     s.tilesKey,
	}

	// change header base on content-type
	ctype := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", ctype)

	err := s.templates.ExecuteTemplate(w, path, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		level.Error(s.logger).Log("msg", "can't execute template", "error", err, "path", path)

		return
	}
}

func isTpl(path string) bool {
	for _, p := range templatesNames {
		if p == path {
			return true
		}
	}

	return false
}
