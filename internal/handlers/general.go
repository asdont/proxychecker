package handlers

import "proxychecker/internal/checker"

type HTTPError struct {
	Error   string `json:"error"`
	Comment string `json:"comment"`
}

const (
	statusNotReady = "NOT_READY"
	statusOk       = "OK"
)

type Checker struct {
	Status  string
	Proxies []checker.ProxyData
}
