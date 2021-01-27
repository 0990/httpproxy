package httpproxy

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func Test_CreateConfig(t *testing.T) {
	cfg := &Config{
		BindAddr:      "0.0.0.0:3128",
		Hosts:         []string{"*"},
		NextProxyAddr: "",
		Verbose:       false,
		PProfAddr:     "0.0.0.0:12580",
	}

	c, _ := json.MarshalIndent(cfg, "", "   ")
	ioutil.WriteFile("config.json", c, 0644)
}
