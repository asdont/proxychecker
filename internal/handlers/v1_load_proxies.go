package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var errProxiesRequired = errors.New("proxies required")

func V1SendProxies() gin.HandlerFunc {
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
