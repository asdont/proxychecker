package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"proxychecker/internal/checker"
)

const shortenUserRequestIDTo = 1e8

var errProxiesRequired = errors.New("proxies required")

type SendProxiesRes struct {
	RequestID string `json:"requestID"`
}

func V1SendProxies(mu *sync.RWMutex, userRequests map[int]Checker, chErr chan<- error) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxies, err := parseProxies(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   err.Error(),
				Comment: "could not read the body",
			})
		}

		if len(proxies) == 0 {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   errProxiesRequired.Error(),
				Comment: "len(proxies) > 0 required",
			})
		}

		userRequestID := createUserRequestID(userRequests)

		mu.Lock()
		userRequests[userRequestID] = Checker{
			Status: statusNotReady,
		}
		mu.Unlock()

		go func() {
			proxiesData, err := checker.CheckProxies(proxies)
			if err != nil {
				chErr <- fmt.Errorf("check proxies: %w", err)
			}

			mu.Lock()
			userRequests[userRequestID] = Checker{
				Status:  statusOk,
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
	userRequestID := time.Now().UnixNano() % shortenUserRequestIDTo

	for {
		if _, exists := userRequests[userRequestID]; exists {
			userRequestID++

			continue
		}

		return userRequestID
	}
}
