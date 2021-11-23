package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"text/template"

	log "github.com/go-kit/log"
	"google.golang.org/grpc/health"

	"github.com/akhenakh/kvtiles/storage"
)

// Server exposes indexes services
type Server struct {
	tileStorage  storage.TileStore
	logger       log.Logger
	appName      string
	healthServer *health.Server
	fileHandler  http.Handler
	templates    *template.Template
	tilesKey     string
}

// New returns a Server
func New(appName, tilesKey string, afs fs.FS, storage storage.TileStore,
	logger log.Logger, healthServer *health.Server) (*Server, error) {
	logger = log.With(logger, "component", "server")

	// static file handler
	fileHandler := http.FileServer(http.FS(afs))

	// computing templates
	pathTpls := make([]string, len(templatesNames))
	for i, name := range templatesNames {
		pathTpls[i] = "./static/" + name
	}
	t, err := template.ParseFiles(pathTpls...)
	if err != nil {
		return nil, fmt.Errorf("can't parse templates: %w", err)
	}

	s := &Server{
		tileStorage:  storage,
		logger:       logger,
		appName:      appName,
		healthServer: healthServer,
		fileHandler:  fileHandler,
		tilesKey:     tilesKey,
		templates:    t,
	}

	return s, nil
}
