package httpproxy

type Config struct {
	BindAddr      string   `json:"bind_addr"`
	LocalDstHosts []string `json:"local_dst_hosts"` //走本地的host

	RemoteAddr string `json:"remote_addr"` //当匹配到远程模式时，会将连接转到这个地址http代理

	Verbose bool `json:"verbose"`
}
