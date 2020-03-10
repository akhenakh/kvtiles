package server

import (
	"net/http"
	"strconv"
	"strings"
)

// ServeHTTP serves the mbtiles for URL such as /tiles/11/618/722.pbf
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sp := strings.Split(req.URL.Path, "/")

	if len(sp) != 5 {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}
	z, _ := strconv.Atoi(sp[2])
	x, _ := strconv.Atoi(sp[3])
	y, _ := strconv.Atoi(strings.Trim(sp[4], ".pbf"))

	data, err := s.tileStorage.ReadTileData(uint8(z), uint64(x), uint64(1<<uint(z)-y-1))
	if err != nil {
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
