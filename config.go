package httpproxy

type Config struct {
	BindAddr string   `json:"bind_addr"`
	Hosts    []string `json:"hosts"` //哪些host走此代理，*代表任意host

	NextProxyAddr string `json:"next_proxy_addr"` //除上面host外的走的http代理

	Verbose bool `json:"verbose"`

	StatsAddr string `json:"stats_addr"` //统计http地址
	PProfAddr string `json:"pprof_addr""`
}
