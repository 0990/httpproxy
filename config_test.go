package httpproxy

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func Test_CreateConfig(t *testing.T) {
	cfg := &Config{
		BindAddr:      "0.0.0.0:8080",
		Hosts:         []string{"baidu.com"},
		NextProxyAddr: "",
		Verbose:       false,
	}

	c, _ := json.MarshalIndent(cfg, "", "   ")
	ioutil.WriteFile("config.json", c, 0644)
}
