package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ip2location/ip2location-go"

	"proxychecker/internal/checker"
	"proxychecker/internal/config"
)

var errProxiesRequired = errors.New("proxies required")

type SendProxiesRes struct {
	RequestID string `json:"requestID"`
}

func V1SendProxies(
	ctx context.Context,
	mu *sync.RWMutex,
	conf config.Conf,
	dbGeo *ip2location.DB,
	userRequests map[int]Checker,
	regexps config.Regexps,
	chErr chan<- error,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawProxies, err := parseProxies(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   err.Error(),
				Comment: "could not read the body",
			})
		}

		if len(rawProxies) == 0 {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   errProxiesRequired.Error(),
				Comment: "len(proxies) > 0 required",
			})
		}

		userRequestID := createUserRequestID(userRequests)

		mu.Lock()
		userRequests[userRequestID] = Checker{
			Status: statusUserRequestNotReady,
		}
		mu.Unlock()

		go func() {
			proxiesData, err := checker.CheckProxies(ctx, conf, dbGeo, rawProxies, regexps)
			if err != nil {
				chErr <- fmt.Errorf("check proxies: %w", err)
			}

			mu.Lock()
			userRequests[userRequestID] = Checker{
				Status:  statusUserRequestOk,
				Proxies: proxiesData,
			}
			mu.Unlock()
		}()

		c.JSON(http.StatusOK, SendProxiesRes{
			RequestID: strconv.Itoa(userRequestID),
		})
	}
}

func parseProxies(c *gin.Context) ([]string, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("io: read all: %w", err)
	}

	return strings.FieldsFunc(string(bytes.TrimSpace(body)), func(r rune) bool {
		return r == '\n'
	}), nil
}

func createUserRequestID(userRequests map[int]Checker) int {
	userRequestID := int(time.Now().UnixNano() % config.ShortenUserRequestIDTo)

	for {
		if _, exists := userRequests[userRequestID]; exists {
			userRequestID++

			continue
		}

		return userRequestID
	}
}
