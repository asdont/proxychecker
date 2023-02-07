package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type GetProxiesURI struct {
	RequestID int `uri:"requestID" binding:"required"`
}

func V1GetProxies(mu *sync.RWMutex, userRequests map[int]Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u GetProxiesURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   err.Error(),
				Comment: ".../api/v1/proxies/[request_id]",
			})

			return
		}

		checkerProxiesData, ok := userRequests[u.RequestID]
		if !ok {
			c.JSON(http.StatusOK, Checker{
				Status: statusUserRequestNotFound,
			})

			return
		}

		if checkerProxiesData.Status == statusUserRequestNotReady {
			c.JSON(http.StatusOK, Checker{
				Status: statusUserRequestNotReady,
			})

			return
		}

		defer func() {
			mu.Lock()
			delete(userRequests, u.RequestID)
			mu.Unlock()
		}()

		c.JSON(http.StatusOK, checkerProxiesData)
	}
}
