package rest

import (
	"context"
	"fmt"
	"isekai-shop/internal/config"
	"isekai-shop/logs"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	app  *echo.Echo
	db   *sqlx.DB
	conf *config.Config
}

func NewServer(conf *config.Config, db *sqlx.DB) Server {
	e := echo.New()

	// Use middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Error: Request timeout",
		Timeout:      conf.Server.Timeout * time.Second,
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: conf.Server.AllowOrigins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.BodyLimit(conf.Server.BodyLimit))

	server := &Server{
		app:  e,
		db:   db,
		conf: conf,
	}

	// Initialize routes
	e.GET("/v1/health", server.healthCheck)

	return *server
}

func (s *Server) Start() {
	url := fmt.Sprintf(":%v", s.conf.Server.Port)

	s.initItemshopRouter()

	// Create done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go s.gracfulShutdown(done)

	err := s.app.Start(url)
	if err != nil && err != http.ErrServerClosed{
		s.app.Logger.Fatalf("Error: %v", err)
	}

	// Wait for graceful shutdown to complete
	<-done
	logs.Info("Graceful shutdown complete")
}

func (s *Server) gracfulShutdown(done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen to interupt signal.
	<-ctx.Done()

	logs.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform server it has 5 second to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.app.Server.Shutdown(ctx); err != nil {
		logs.Info("Server force to shutdown with error: ", zap.Error(err))
	}

	logs.Info("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func (s *Server) healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
