package checker

type ProxyData struct {
	Address string `json:"address"`
	RealIP  string `json:"realIP"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Comment string `json:"comment,omitempty"`
	Error   string `json:"error,omitempty"`
}

func CheckProxies(proxyList []string) ([]ProxyData, error) {
	// TODO
	return nil, nil
}
