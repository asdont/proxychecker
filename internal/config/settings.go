package config

import "regexp"

const (
	FileConf  = "configs/conf.toml"
	FileGeoDB = "data/geo-db.bin"
)

const ShortenUserRequestIDTo = 1e8

const DelayBetweenProxyChecksMs = 50

const ServerMaxHeaderBytesMib = 1 << 20

const (
	CheckForLostUserRequestsEveryMinutes = 10
	DeleteLostUserRequestsAfterSeconds   = 60
)

type ProxyProtocol string

const (
	ProxyStatusOk   = "OK"
	ProxyStatusFail = "FAIL"
)

const (
	regexpProxyScheme   = `(https|http|socks4|socks5):`
	regexpProxyAuth     = `(@|/).+:.+$`
	regexpProxyHostPort = `(\d{1,3}[.]){3}\d{1,3}:\d{0,5}`
	regexpIPv4          = `(\d{1,3}[.]){3}\d{1,3}`
	regexpIPv6          = `([0-9A-Fa-f]{1,4})(:[0-9A-Fa-f]{1,4})*::([0-9A-Fa-f]{1,4})(:[0-9A-Fa-f]{1,4})*|([0-9A-Fa-f]{1,4})(:[0-9A-Fa-f]{1,4}){7}` //nolint:lll
)

type Regexps struct {
	ProxyScheme   *regexp.Regexp
	ProxyAuth     *regexp.Regexp
	ProxyHostPort *regexp.Regexp
	IPv4          *regexp.Regexp
	IPv6          *regexp.Regexp
}

func CompileRegexps() Regexps {
	return Regexps{
		ProxyScheme:   regexp.MustCompile(regexpProxyScheme),
		ProxyAuth:     regexp.MustCompile(regexpProxyAuth),
		ProxyHostPort: regexp.MustCompile(regexpProxyHostPort),
		IPv4:          regexp.MustCompile(regexpIPv4),
		IPv6:          regexp.MustCompile(regexpIPv6),
	}
}
