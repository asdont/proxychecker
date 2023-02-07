package handlers

import "proxychecker/internal/checker"

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
}
