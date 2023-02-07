package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"proxychecker/internal/config"
)

func TestConvertProxies(t *testing.T) {
	cases := []struct {
		name            string
		rawProxies      []string
		expectedProxies []string
	}{
		{
			name: "positive_proxies",
			rawProxies: []string{
				"http://1.1.1.1:1",
				"https://10.10.10.20:20",
				"socks4://100.100.100.300:300",
				"socks5://1.20.300.4:4000",
			},
			expectedProxies: []string{
				"http://1.1.1.1:1",
				"https://10.10.10.20:20",
				"socks4://100.100.100.300:300",
				"socks5://1.20.300.4:4000",
			},
		},
		{
			name: "positive_proxies_with_auth",
			rawProxies: []string{
				"http://login:pass@1.1.1.1:2",
				"https://login:pass@10.10.10.20:20",
				"socks4://login:pass@100.100.100.300:200",
				"socks5://login:pass@1.20.300.4:2000",
			},
			expectedProxies: []string{
				"http://login:pass@1.1.1.1:2",
				"https://login:pass@10.10.10.20:20",
				"socks4://login:pass@100.100.100.300:200",
				"socks5://login:pass@1.20.300.4:2000",
			},
		},
		{
			name: "negative_proxies",
			rawProxies: []string{
				"http:// 1.1.1.1:1",
				"htt ://10. 10.10.20:20",
				"socks4://100.100.100:30",
				"socks4://100.100.100.100",
				"socks://1.20.300.4:4000",
			},
			expectedProxies: []string{
				"http://1.1.1.1:1",
				"",
				"",
				"",
				"",
			},
		},
		{
			name: "negative_proxies_with_auth",
			rawProxies: []string{
				"http:// login:pass  @ 1.1.1.1:1",
				"htt ://login:pass@10. 10.10.20:20",
				"socks4://login:pass@100.100.100:30",
				"socks4://pass@100.100.100.100",
				"socks://login@1.20.300.4:4000",
			},
			expectedProxies: []string{
				"http://login:pass@1.1.1.1:1",
				"",
				"",
				"",
				"",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			convertProxies(tt.rawProxies, config.CompileRegexps())

			assert.Equal(t, tt.expectedProxies, tt.rawProxies)
		})
	}
}
