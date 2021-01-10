package httpproxy

import (
	"golang.org/x/net/idna"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"unicode/utf8"
)

var hasPort = regexp.MustCompile(`:\d+$`)

var portMap = map[string]string{
	"http":   "80",
	"https":  "443",
	"socks5": "1080",
}

func canonicalAddr(url *url.URL) string {
	addr := url.Hostname()
	if v, err := idnaASCII(addr); err == nil {
		addr = v
	}
	port := url.Port()
	if port == "" {
		port = portMap[url.Scheme]
	}
	return net.JoinHostPort(addr, port)
}

func idnaASCII(v string) (string, error) {
	if isASCII(v) {
		return v, nil
	}
	return idna.Lookup.ToASCII(v)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func hostport(r *http.Request) string {
	if r.Method == "CONNECT" {
		host := r.URL.Host
		if !hasPort.MatchString(host) {
			host += ":80"
		}
		return host
	} else {
		return canonicalAddr(r.URL)
	}
}
