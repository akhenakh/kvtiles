package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// _ "net/http/pprof"

	log "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/akhenakh/kvtiles/loglevel"
	"github.com/akhenakh/kvtiles/server"
	pstorage "github.com/akhenakh/kvtiles/storage/pogreb"
)

const appName = "kvtilesd"

var (
	version = "no version from LDFLAGS"

	logLevel        = flag.String("logLevel", "INFO", "DEBUG|INFO|WARN|ERROR")
	dbPath          = flag.String("dbPath", "map.db", "Database path")
	httpMetricsPort = flag.Int("httpMetricsPort", 8088, "http port")
	httpAPIPort     = flag.Int("httpAPIPort", 8080, "http API port")
	healthPort      = flag.Int("healthPort", 6666, "grpc health port")
	tilesKey        = flag.String("tilesKey", "", "A key to protect your tiles access")
	allowOrigin     = flag.String("allowOrigin", "*", "Access-Control-Allow-Origin")

	//go:embed static/*
	staticFS embed.FS

	httpServer        *http.Server
	grpcHealthServer  *grpc.Server
	httpMetricsServer *http.Server
)

func main() {
	flag.Parse()

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "caller", log.Caller(5), "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "app", appName)
	logger = loglevel.NewLevelFilterFromString(logger, *logLevel)

	stdlog.SetOutput(log.NewStdlibAdapter(logger))

	level.Info(logger).Log("msg", "Starting app", "version", version)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// catch termination
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	// pprof
	// go func() {
	// 	stdlog.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	storage, clean, err := pstorage.NewStorage(*dbPath, logger)
	if err != nil {
		level.Error(logger).Log("msg", "can't open storage for writing", "error", err)
		os.Exit(2)
	}
	defer clean()

	infos, ok, err := storage.LoadMapInfos()
	if err != nil {
		level.Error(logger).Log("msg", "failed to read infos", "error", err)
		os.Exit(2)
	}
	if !ok {
		level.Error(logger).Log("msg", "no map infos")
		os.Exit(2)
	}

	// gRPC Health Server
	healthServer := health.NewServer()
	g.Go(func() error {
		grpcHealthServer = grpc.NewServer()

		healthpb.RegisterHealthServer(grpcHealthServer, healthServer)

		haddr := fmt.Sprintf(":%d", *healthPort)
		hln, err := net.Listen("tcp", haddr)
		if err != nil {
			level.Error(logger).Log("msg", "gRPC Health server: failed to listen", "error", err)
			os.Exit(2)
		}
		level.Info(logger).Log("msg", fmt.Sprintf("gRPC health server listening at %s", haddr))

		return grpcHealthServer.Serve(hln)
	})

	// server
	server, err := server.New(appName, *tilesKey, staticFS, storage, logger, healthServer)
	if err != nil {
		level.Error(logger).Log("msg", "can't get a working server", "error", err)
		os.Exit(2)
	}

	// web server metrics
	g.Go(func() error {
		httpMetricsServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", *httpMetricsPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		level.Info(logger).Log("msg", fmt.Sprintf("HTTP Metrics server listening at :%d", *httpMetricsPort))

		versionGauge.WithLabelValues(version).Add(1)
		dataVersionGauge.WithLabelValues(
			fmt.Sprintf("%s %s", infos.Region, infos.IndexTime.Format(time.RFC3339)),
		).Add(1)

		// Register Prometheus metrics handler.
		http.Handle("/metrics", promhttp.Handler())

		if err := httpMetricsServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	// web server
	g.Go(func() error {
		// metrics middleware.
		metricsMwr := middleware.New(middleware.Config{
			Recorder: metrics.NewRecorder(metrics.Config{Prefix: appName}),
		})

		r := mux.NewRouter()

		r.Handle("/tiles/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.pbf", std.Handler("/tiles/", metricsMwr, server))

		// serving templates and static files
		r.PathPrefix("/static/").HandlerFunc(server.StaticHandler)

		r.HandleFunc("/healthz", server.HealthHandler)

		r.HandleFunc("/version", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			m := map[string]interface{}{"version": version, "infos": infos}
			b, _ := json.Marshal(m)
			w.Write(b)
		})

		httpServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", *httpAPIPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler: handlers.CORS(
				handlers.AllowedOrigins([]string{*allowOrigin}),
				handlers.AllowedMethods([]string{"GET"}))(r),
		}
		level.Info(logger).Log("msg", fmt.Sprintf("HTTP API server listening at :%d", *httpAPIPort))

		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	healthServer.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", appName), healthpb.HealthCheckResponse_SERVING)
	level.Info(logger).Log("msg", "serving status to SERVING")

	select {
	case <-interrupt:
		cancel()

		break
	case <-ctx.Done():
		break
	}

	level.Warn(logger).Log("msg", "received shutdown signal")

	healthServer.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", appName), healthpb.HealthCheckResponse_NOT_SERVING)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if httpMetricsServer != nil {
		_ = httpMetricsServer.Shutdown(shutdownCtx)
	}

	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}

	if grpcHealthServer != nil {
		grpcHealthServer.GracefulStop()
	}

	err = g.Wait()
	if err != nil {
		level.Error(logger).Log("msg", "server returning an error", "error", err)
		os.Exit(2)
	}
}
