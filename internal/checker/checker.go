package checker

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ip2location/ip2location-go"

	"proxychecker/internal/config"
)

type ProxyData struct {
	Status  string `json:"status"`
	Address string `json:"address"`
	RealIP  string `json:"realIP"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Comment string `json:"comment,omitempty"`
}

func CheckProxies(
	ctx context.Context,
	conf config.Conf,
	dbGeo *ip2location.DB,
	rawProxies []string,
	regexps config.Regexps,
) ([]ProxyData, error) {
	proxiesData := make([]ProxyData, 0, len(rawProxies))

	chSuccess := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	var wg sync.WaitGroup

	for _, proxyAddress := range convertProxies(rawProxies, regexps) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			proxyData, err := checkProxy(
				ctx, dbGeo, regexps, proxyAddress,
				conf.Checker.RequestTimeoutSeconds, conf.Checker.ServiceMyIP, conf.Checker.HeaderUserAgent,
			)
			if err != nil {
				chErr <- fmt.Errorf("check proxy: %w", err)
			}

			proxiesData = append(proxiesData, proxyData)
		}()

		time.Sleep(time.Millisecond * config.DelayBetweenProxyChecksMs)
	}

	go func() {
		wg.Wait()

		chSuccess <- struct{}{}
	}()

	select {
	case err := <-chErr:
		return nil, err

	case <-chSuccess:
		return proxiesData, nil
	}
}

func checkProxy(
	ctx context.Context,
	dbGeo *ip2location.DB,
	regexps config.Regexps,
	proxyAddress string,
	requestTimeoutSeconds int,
	uriServiceMyIP string,
	userAgentHeader string,
) (ProxyData, error) {
	proxyData := ProxyData{
		Address: proxyAddress,
		Status:  config.ProxyStatusFail,
	}

	proxyURL, err := url.Parse(proxyAddress)
	if err != nil {
		proxyData.Comment = err.Error()

		return proxyData, nil //nolint:nilerr
	}

	client, err := createClient(proxyURL)
	if err != nil {
		return ProxyData{}, fmt.Errorf("create transport: %w", err)
	}

	pageBodyMyIP, err := getPageBodyMyIP(ctx, client, requestTimeoutSeconds, uriServiceMyIP, userAgentHeader)
	if err != nil {
		proxyData.Comment = err.Error()

		if errors.Is(err, errProxyFailed) || errors.Is(err, errStatusCode) {
			return proxyData, nil
		}

		return ProxyData{}, fmt.Errorf("get page body: my ip: %w", err)
	}

	proxyData.Status = config.ProxyStatusOk
	proxyData.RealIP = extractRealIP(pageBodyMyIP, regexps)

	ipData, err := dbGeo.Get_all(proxyData.RealIP)
	if err != nil {
		return ProxyData{}, fmt.Errorf("db geo: get all: %w", err)
	}

	proxyData.Country = ipData.Country_long
	proxyData.Region = ipData.Region
	proxyData.City = ipData.City

	return proxyData, nil
}

func convertProxies(rawProxies []string, regexps config.Regexps) []string {
	for i, rawProxy := range rawProxies {
		rawProxy = strings.Join(strings.Fields(rawProxy), "")

		hostPort := regexps.ProxyHostPort.FindString(rawProxy)
		if hostPort == "" {
			rawProxies[i] = ""

			continue
		}

		rawProxy = strings.ReplaceAll(rawProxy, hostPort, "")

		scheme := strings.TrimRight(regexps.ProxyScheme.FindString(rawProxy), ":")
		if scheme == "" {
			rawProxies[i] = ""

			continue
		}

		rawProxy = strings.ReplaceAll(rawProxy, scheme, "")

		auth := regexps.ProxyAuth.FindString(rawProxy)
		auth = strings.TrimLeft(auth, "/@")
		auth = strings.TrimRight(auth, "@")

		var sb strings.Builder

		sb.WriteString(scheme)
		sb.WriteString("://")

		if auth != "" {
			sb.WriteString(auth)
			sb.WriteString("@")
		}

		sb.WriteString(hostPort)

		rawProxies[i] = sb.String()
	}

	return rawProxies
}

func extractRealIP(body []byte, regexps config.Regexps) string {
	realIP4 := regexps.IPv4.Find(body)
	if realIP4 != nil {
		return strings.TrimSpace(string(realIP4))
	}

	realIP6 := regexps.IPv6.Find(body)
	if realIP6 != nil {
		return strings.TrimSpace(string(realIP6))
	}

	return ""
}
