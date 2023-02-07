package checker

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

var (
	errProxyFailed        = errors.New("proxy failed")
	errStatusCode         = errors.New("status code")
	errUnknownProxyScheme = errors.New("unknown proxy scheme")
)

func createClient(proxyURL *url.URL) (*http.Client, error) {
	client := new(http.Client)

	switch proxyURL.Scheme {
	case "http", "https":
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			ProxyConnectHeader: http.Header{
				"Proxy-Authorization": []string{
					"Basic " + base64.StdEncoding.EncodeToString([]byte(proxyURL.User.String())),
				},
			},
		}

	case "socks5", "socks4":
		auth := new(proxy.Auth)

		password, ok := proxyURL.User.Password()
		if ok {
			auth.User = proxyURL.User.Username()
			auth.Password = password
		}

		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("socks5: conn failed: %w", err)
		}

		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}

		client.Transport = &http.Transport{
			DialContext: dialContext,
		}

	default:
		return nil, fmt.Errorf("scheme: %s: %w", proxyURL.Scheme, errUnknownProxyScheme)
	}

	return client, nil
}

func getPageBodyMyIP(
	ctx context.Context,
	client *http.Client,
	requestTimeoutSeconds int,
	uriServiceMyIP string,
	userAgentHeader string,
) ([]byte, error) {
	ctxRequest, cancel := context.WithTimeout(ctx, time.Second*time.Duration(requestTimeoutSeconds))
	defer cancel()

	req, err := http.NewRequestWithContext(ctxRequest, http.MethodGet, uriServiceMyIP, nil)
	if err != nil {
		return nil, fmt.Errorf("http: new request: %w", err)
	}

	req.Header.Set("User-Agent", userAgentHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: do: %w: %v", errProxyFailed, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("close body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status-code: %d: %w", resp.StatusCode, errStatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}
