package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"proxychecker/internal/handlers"
)

var errWrongParameter = errors.New("wrong parameter")

type ServerHTTP struct {
	Mode            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownMaxTime time.Duration
}

func ConfigHTTP(
	mode string,
	port int,
	readTimeoutSeconds int,
	writeTimeoutSeconds int,
	shutdownMaxTimeSeconds int,
) (ServerHTTP, error) {
	if mode == "" {
		return ServerHTTP{}, fmt.Errorf("mode: %w", errWrongParameter)
	}

	if port == 0 {
		return ServerHTTP{}, fmt.Errorf("port: %w", errWrongParameter)
	}

	if readTimeoutSeconds == 0 {
		return ServerHTTP{}, fmt.Errorf("read timeout: %w", errWrongParameter)
	}

	if writeTimeoutSeconds == 0 {
		return ServerHTTP{}, fmt.Errorf("write timout: %w", errWrongParameter)
	}

	if shutdownMaxTimeSeconds == 0 {
		return ServerHTTP{}, fmt.Errorf("shutdown max time: %w", errWrongParameter)
	}

	return ServerHTTP{
		Mode:            gin.ReleaseMode,
		Port:            strconv.Itoa(port),
		ReadTimeout:     time.Second * time.Duration(readTimeoutSeconds),
		WriteTimeout:    time.Second * time.Duration(writeTimeoutSeconds),
		ShutdownMaxTime: time.Second * time.Duration(shutdownMaxTimeSeconds),
	}, nil
}

func (s ServerHTTP) Run(mu *sync.RWMutex, userRequests map[int]handlers.Checker, chErr chan<- error) error {
	gin.SetMode(s.Mode)

	router := gin.New()

	router.Use(
		gin.Recovery(),
	)

	setRoutes(mu, router, userRequests, chErr)

	srv := &http.Server{
		Addr:         ":" + s.Port,
		Handler:      router,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	go func() {
		if err := stopServer(srv, s.ShutdownMaxTime); err != nil {
			chErr <- fmt.Errorf("stop http server: %w", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func setRoutes(mu *sync.RWMutex, router *gin.Engine, userRequests map[int]handlers.Checker, chErr chan<- error) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/proxies", handlers.V1SendProxies(mu, userRequests, chErr))
	}
}

func stopServer(srv *http.Server, shutdownMaxTime time.Duration) error {
	chSig := make(chan os.Signal, 1)

	signal.Notify(chSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-chSig

	ctx, cancel := context.WithTimeout(context.Background(), shutdownMaxTime)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("shutdown: forced stop: %w", err)
		}

		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
