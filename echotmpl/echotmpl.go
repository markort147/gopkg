package echotmpl

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

type Config struct {
	Port            int
	LogOutputPath   io.Writer
	LogLevel        log.Lvl
	DefLogger       log.Logger
	FileSystem      fs.FS
	IndexPath       string
	TemplatesPath   string
	RoutesRegister  func(e *Echo)
	ShutdownSignals []os.Signal
	CustomFuncs     FuncMap
}

type Echo = echo.Echo
type FuncMap = template.FuncMap

// StartServer initializes the echo server
// and registers all the endpoints
func StartServer(cfg *Config) (*sync.WaitGroup, error) {
	// verify the configuration
	if err := fixConfig(cfg); err != nil {
		return nil, fmt.Errorf("server configuration error: %w", err)
	}

	// create the server and configure the middleware
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Output: cfg.LogOutputPath}))
	e.Logger.SetLevel(cfg.LogLevel)
	e.Use(middleware.Recover())

	// serve index and register custom routes
	e.Renderer = newTemplateRenderer(cfg.FileSystem, cfg.TemplatesPath, cfg.CustomFuncs)
	e.FileFS("/", cfg.IndexPath, cfg.FileSystem)
	if cfg.RoutesRegister != nil {
		cfg.RoutesRegister(e)
	}

	// intercept shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, cfg.ShutdownSignals...)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-quit
		cfg.DefLogger.Info("Shutting down the server")

		// Create a context with a timeout to allow for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server
		if err := e.Shutdown(ctx); err != nil {
			cfg.DefLogger.Error("Server forced to shutdown: ", err)
		}

		cfg.DefLogger.Info("Server exiting")
	}()

	// start the server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("Error starting the server: ", err)
		}
	}()

	return &wg, nil
}

func fixConfig(cfg *Config) error {
	if cfg.FileSystem == nil {
		return fmt.Errorf("no filesystem configuration")
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.LogOutputPath == nil {
		cfg.LogOutputPath = os.Stdout
	}
	if cfg.LogLevel == 0 {
		cfg.LogLevel = log.INFO
	}
	if cfg.IndexPath == "" {
		cfg.IndexPath = "assets/index.html"
	}
	if cfg.TemplatesPath == "" {
		cfg.TemplatesPath = "assets/templates/*"
	}
	if cfg.ShutdownSignals == nil || len(cfg.ShutdownSignals) == 0 {
		cfg.ShutdownSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	return nil
}

type templateRenderer struct {
	tmpl *template.Template
}

func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

func newTemplateRenderer(fs fs.FS, path string, funcMap template.FuncMap) echo.Renderer {
	tmpl := template.New("templates")
	if funcMap != nil {
		tmpl.Funcs(funcMap)
	}
	return &templateRenderer{
		tmpl: template.Must(tmpl.ParseFS(fs, path)),
	}
}
