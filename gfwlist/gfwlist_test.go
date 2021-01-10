package gfwlist

import (
	"log"
	"net/url"
	"testing"
)

var tests = []string{
	"http://www.ggogle.com",
	"http://www.google.com",
	"https://www.facebook.com",
	"http://www.qq.com",
	"https://cn-proxy.com",
	"http://x.ftchinese.com/story",
	"https://www.ftchinese.com/story",
	"https://galenwu.com",
	"http://xx.sf.net",
	"https://static.xx.fbcdn.net",
}
var testDomains = []string{
	"www.ggogle.com",
	"www.google.com",
	"facebook.com",
	"www.qq.com",
	"cn-proxy.com",
	"x.ftchinese.com",
	"zz.ftchinese.com",
	"galenwu.com",
	"xx.sf.net",
	"boxunx.azurewebsites.net",
	"fbcdn1.akamaihd.net",
	"xx.google.xx",
	"twimg.edgesuite.net",
	"static.xx.fbcdn.net",
}

var testDomains1 = []string{
	"static.bbc.co.uk",
	"on.wsj.com",
	"api.nhk.or.jp",
}

func TestGFWList(t *testing.T) {
	gfw, _ := NewFromFile("gfwlist.txt", false)
	for _, rule := range tests {
		u, _ := url.Parse(rule)
		_, r2 := gfw.Match(u)
		log.Printf("%s->%v", rule, r2)
	}
}

func TestGFWListDomain(t *testing.T) {
	gfw, _ := NewFromFile("gfwlist.txt", false)
	for _, rule := range testDomains {
		log.Printf("%s->%v", rule, gfw.IsDomainBlocked(rule))
	}
}

func TestGFWListDomain1(t *testing.T) {
	gfw, _ := NewFromFile("gfwlist.txt", true)
	for _, rule := range testDomains1 {
		log.Printf("%s->%v", rule, gfw.IsDomainBlocked(rule))
	}
}
