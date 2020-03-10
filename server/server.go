package server

import (
	log "github.com/go-kit/kit/log"
	"google.golang.org/grpc/health"

	"github.com/akhenakh/kvtiles/storage"
)

// Server exposes indexes services
type Server struct {
	tileStorage  storage.TileStore
	logger       log.Logger
	healthServer *health.Server
}

// New returns a Server
func New(storage storage.TileStore,
	logger log.Logger, healthServer *health.Server) (*Server, error) {
	logger = log.With(logger, "component", "server")

	s := &Server{
		tileStorage:  storage,
		logger:       logger,
		healthServer: healthServer,
	}

	return s, nil
}
