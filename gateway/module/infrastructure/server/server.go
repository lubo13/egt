package main

import (
	"context"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Route interface {
	http.Handler
	Pattern() string
}

// NewHTTPServer sets up the HTTP server with graceful shutdown.
func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger) *http.Server {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			log.Info("Starting HTTP server", zap.String("addr", srv.Addr))

			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP server error: ", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			log.Info("Shutting down HTTP server gracefully")

			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error("HTTP server shutdown error", zap.Error(err))
				return err
			}
			log.Info("HTTP server stopped")
			return nil
		},
	})

	return srv
}

// NewServeMux registers all routes into a single http.ServeMux.
func NewServeMux(routes []Route) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route.Pattern(), route)
	}

	return mux
}
