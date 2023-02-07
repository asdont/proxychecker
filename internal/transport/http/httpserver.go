package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ip2location/ip2location-go"

	"proxychecker/internal/config"
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

func (s ServerHTTP) Run(
	ctx context.Context,
	mu *sync.RWMutex,
	conf config.Conf,
	dbGeo *ip2location.DB,
	userRequests map[int]handlers.Checker,
	regexps config.Regexps,
	chErr chan<- error,
) error {
	gin.SetMode(s.Mode)

	router := gin.New()

	router.Use(
		gin.Recovery(),
	)

	setRoutes(ctx, mu, conf, dbGeo, router, userRequests, regexps, chErr)

	srv := &http.Server{
		Addr:           ":" + s.Port,
		Handler:        router,
		ReadTimeout:    s.ReadTimeout,
		WriteTimeout:   s.WriteTimeout,
		MaxHeaderBytes: config.ServerMaxHeaderBytesMib,
	}

	go func() {
		if err := stopServer(ctx, srv, s.ShutdownMaxTime); err != nil {
			chErr <- fmt.Errorf("stop http server: %w", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func setRoutes(
	ctx context.Context,
	mu *sync.RWMutex,
	conf config.Conf,
	dbGeo *ip2location.DB,
	router *gin.Engine,
	userRequests map[int]handlers.Checker,
	regexps config.Regexps,
	chErr chan<- error,
) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/proxies", handlers.V1SendProxies(ctx, mu, conf, dbGeo, userRequests, regexps, chErr))
		v1.GET("/proxies/:requestID", handlers.V1GetProxies(mu, userRequests))
	}
}

func stopServer(ctx context.Context, srv *http.Server, shutdownMaxTime time.Duration) error {
	<-ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownMaxTime)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("shutdown: forced stop: %w", err)
		}

		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
