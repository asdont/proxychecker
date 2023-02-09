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

var errPartial = errors.New("partial error")

type ProxyData struct {
	Status  string `json:"status"`
	Address string `json:"address"`
	RealIP  string `json:"realIP"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Comment string `json:"comment"`
}

func CheckProxies(
	ctx context.Context,
	conf config.Conf,
	dbGeo *ip2location.DB,
	rawProxies []string,
	regexps config.Regexps,
) ([]*ProxyData, error) {
	chSuccess := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	proxiesData := convertProxies(rawProxies, regexps)

	var wg sync.WaitGroup

	for _, proxyData := range proxiesData {
		if proxyData.Comment != "" { // Comment = error
			continue
		}

		wg.Add(1)

		go func(proxyData *ProxyData) {
			defer wg.Done()

			if err := checkProxyAndSetData(
				ctx, dbGeo, regexps, proxyData,
				conf.Checker.RequestTimeoutSeconds, conf.Checker.ServiceMyIP, conf.Checker.HeaderUserAgent,
			); err != nil && !errors.Is(err, errPartial) {
				chErr <- fmt.Errorf("check proxy: %w", err)
			}
		}(proxyData)

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

func checkProxyAndSetData(
	ctx context.Context,
	dbGeo *ip2location.DB,
	regexps config.Regexps,
	proxyData *ProxyData,
	requestTimeoutSeconds int,
	uriServiceMyIP string,
	userAgentHeader string,
) error {
	proxyURL, err := url.Parse(proxyData.Address)
	if err != nil {
		proxyData.Status = config.ProxyStatusFail
		proxyData.Comment = "wrong proxy address"

		return fmt.Errorf("url: parse: %w: %v", errPartial, err)
	}

	client, err := createClient(proxyURL)
	if err != nil {
		proxyData.Status = config.ProxyStatusFail
		proxyData.Comment = "client creation error"

		return fmt.Errorf("create client: %w: %v", errPartial, err)
	}

	pageBodyMyIP, err := getPageBodyMyIP(ctx, client, requestTimeoutSeconds, uriServiceMyIP, userAgentHeader)
	if err != nil {
		proxyData.Status = config.ProxyStatusFail
		proxyData.Comment = "bad proxy"

		if errors.Is(err, errProxyFailed) || errors.Is(err, errStatusCode) {
			return nil
		}

		return fmt.Errorf("get page body: my ip: %w", err)
	}

	proxyData.Status = config.ProxyStatusOk
	proxyData.RealIP = extractRealIP(pageBodyMyIP, regexps)

	ipData, err := dbGeo.Get_all(proxyData.RealIP)
	if err != nil {
		return fmt.Errorf("db geo: get all: %w", err)
	}

	proxyData.Country = ipData.Country_long
	proxyData.Region = ipData.Region
	proxyData.City = ipData.City

	return nil
}

func convertProxies(rawProxies []string, regexps config.Regexps) []*ProxyData {
	proxies := make([]*ProxyData, len(rawProxies))

	for i, rawProxy := range rawProxies {
		proxyData := &ProxyData{
			Address: rawProxy,
		}

		rawProxy = strings.Join(strings.Fields(rawProxy), "")

		hostPort := regexps.ProxyHostPort.FindString(rawProxy)

		rawProxy = strings.ReplaceAll(rawProxy, hostPort, "")

		scheme := strings.TrimRight(regexps.ProxyScheme.FindString(rawProxy), ":")

		rawProxy = strings.ReplaceAll(rawProxy, scheme, "")

		auth := regexps.ProxyAuth.FindString(rawProxy)
		auth = strings.TrimLeft(auth, "/@")
		auth = strings.TrimRight(auth, "@")

		if hostPort == "" || scheme == "" {
			proxyData.Comment = "wrong host or port"

			proxies[i] = proxyData
		} else {
			proxyData.Address = createProxyAddress(scheme, auth, hostPort)

			proxies[i] = proxyData
		}
	}

	return proxies
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

func createProxyAddress(scheme, auth, hostPort string) string {
	var sb strings.Builder

	sb.WriteString(scheme)
	sb.WriteString("://")

	if auth != "" {
		sb.WriteString(auth)
		sb.WriteString("@")
	}

	sb.WriteString(hostPort)

	return sb.String()
}
