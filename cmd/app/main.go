package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"proxychecker/internal/config"
	"proxychecker/internal/transport/http"
)

const fileConf = "conf.toml"

func main() {
	chSuccess := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	conf, err := config.GetConfig(fileConf)
	if err != nil {
		log.Fatalf("get config: %v", err)
	}

	serverHTTP, err := http.ConfigHTTP(gin.ReleaseMode, conf.Server.Port,
		conf.Server.ReadTimeoutSeconds, conf.Server.WriteTimeoutSeconds, conf.Server.ShutdownMaxTimeSeconds)
	if err != nil {
		log.Fatalf("http config: %v", err)
	}

	go func() {
		if err := serverHTTP.Run(chErr); err != nil {
			log.Fatalf("server: http: %v", err)
		}
	}()

	select {
	case <-chSuccess:
	case err := <-chErr:
		log.Fatal(err)
	}
}
