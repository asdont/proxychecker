package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"

	"proxychecker/internal/app"
	"proxychecker/internal/config"
	"proxychecker/internal/handlers"
	"proxychecker/internal/transport/http"

	_ "proxychecker/docs"
)

// @title API Proxy Checker
// @version 1.0

// @host 127.0.0.1:30122
// @schemes http
// @BasePath /api
// @query.collection.format multi
func main() {
	var mu sync.RWMutex

	chSuccess := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	userRequests := make(map[int]handlers.Checker)

	conf, err := config.GetConfig(config.FileConf)
	if err != nil {
		log.Fatalf("get config: %v", err)
	}

	dbGeo, err := app.DBGeo(config.FileGeoDB)
	if err != nil {
		log.Fatalf("geo db: %v", err)
	}

	serverHTTP, err := http.ConfigHTTP(gin.ReleaseMode, conf.Server.Port,
		conf.Server.ReadTimeoutSeconds, conf.Server.WriteTimeoutSeconds, conf.Server.ShutdownMaxTimeSeconds)
	if err != nil {
		log.Fatalf("http config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		appStop()

		cancel()
	}()

	go func() {
		log.Printf("[SERVER][HTTP] port: %d\n", conf.Server.Port)

		if err := serverHTTP.Run(ctx, &mu, conf, dbGeo, userRequests, config.CompileRegexps(), chErr); err != nil {
			log.Fatalf("server: http: %v", err)
		}

		chSuccess <- struct{}{}
	}()

	select {
	case <-chSuccess:
		log.Println("server closed... ok")

	case err := <-chErr:
		log.Fatal(err) //nolint:gocritic
	}
}

func appStop() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	log.Println("signal:", <-sigChan)
}
