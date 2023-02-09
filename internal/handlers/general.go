package handlers

import (
	"sync"
	"time"

	"proxychecker/internal/checker"
	"proxychecker/internal/config"
)

type HTTPError struct {
	Error   string `json:"error"`
	Comment string `json:"comment"`
}

const (
	statusUserRequestNotFound = "NOT_FOUND"
	statusUserRequestNotReady = "NOT_READY"
	statusUserRequestOk       = "OK"
)

type Checker struct {
	Status  string              `json:"status"`
	Proxies []checker.ProxyData `json:"proxies,omitempty"`
	Added   time.Time           `json:"-"`
}

// DeleteLostUserRequests delete requests that the user has not completed.
func DeleteLostUserRequests(mu *sync.RWMutex, userRequests map[int]Checker) {
	for {
		time.Sleep(time.Minute * config.CheckForLostUserRequestsEveryMinutes)

		deleteLostUserRequests(mu, userRequests, config.DeleteLostUserRequestsAfterSeconds)
	}
}

func deleteLostUserRequests(mu *sync.RWMutex, userRequests map[int]Checker, deleteAfterSeconds int) {
	for requestID, checkerData := range userRequests {
		if time.Since(checkerData.Added).Seconds() > float64(deleteAfterSeconds) {
			mu.Lock()
			delete(userRequests, requestID)
			mu.Unlock()
		}
	}
}
