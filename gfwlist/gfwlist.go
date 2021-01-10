package gfwlist

import (
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/publicsuffix"
)

type hostRule struct {
	pattern string
}

func (r *hostRule) matchDomain(domain string) bool {
	if strings.Contains(r.pattern, "*") {
		if r.pattern[0] == '.' {
			r.pattern = "*" + r.pattern
		}
		matched, _ := filepath.Match(r.pattern, domain)
		return matched
	}
	if r.pattern[0] == '.' {
		return strings.HasSuffix(domain, r.pattern)
	}
	domain, _ = publicsuffix.EffectiveTLDPlusOne(domain)
	return domain == r.pattern
	//return strings.Contains(domain, r.pattern)
}

func (r *hostRule) match(u *url.URL) bool {
	return r.matchDomain(u.Host)
}

type urlWildcardRule struct {
	pattern     string
	prefixMatch bool
	suffixMatch bool
	regex       *regexp.Regexp
}

func (r *urlWildcardRule) init() {
	if strings.Contains(r.pattern, "*") {
		pattern := r.pattern
		pattern = strings.Replace(pattern, ".", "\\.", -1)
		pattern = strings.Replace(pattern, "*", ".*", -1)
		if r.prefixMatch {
			pattern = "^" + pattern
		}
		if r.suffixMatch {
			pattern = pattern + "$"
		}
		reg, err := regexp.Compile(pattern)
		if nil != err {
			log.Printf("Invalid regexp rule:%s with reason:%v", pattern, err)
		}
		r.regex = reg
	}
}

func (r *urlWildcardRule) match(u *url.URL) bool {
	// if len(u.Scheme) == 0 {
	// 	u.Scheme = "https"
	// }
	if nil == r.regex {
		if r.prefixMatch {
			return strings.HasPrefix(u.String(), r.pattern)
		}
		if r.suffixMatch {
			return strings.HasSuffix(u.String(), r.pattern)
		}
		return strings.Contains(u.String(), r.pattern)
	}
	return r.regex.MatchString(u.String())
}

type regexRule struct {
	pattern string
}

func (r *regexRule) match(u *url.URL) bool {
	// if len(req.URL.Scheme) == 0 {
	// 	req.URL.Scheme = "https"
	// }
	matched, err := regexp.MatchString(r.pattern, u.String())
	if nil != err {
		log.Printf("Invalid regex pattern:%s wiuth reason:%v", r.pattern, err)
	}
	return matched
}

type whiteListRule struct {
	r gfwListRule
}

func (r *whiteListRule) match(u *url.URL) bool {
	return r.r.match(u)
}

type gfwListRule interface {
	match(u *url.URL) bool
}

type domainRule struct {
	domain hostRule
	rule   gfwListRule
}

type GFWList struct {
	fastMatchMap      map[string][]gfwListRule
	domainPattenRules []domainRule
	ruleList          []gfwListRule
	mutex             sync.Mutex
}

func (gfw *GFWList) parse(line string) error {
	str := strings.TrimSpace(string(line))
	//skip comment & empty line
	if strings.HasPrefix(str, "!") || len(str) == 0 || strings.HasPrefix(str, "[") {
		return nil
	}
	var rule gfwListRule
	isWhileListRule := false
	if strings.HasPrefix(str, "@@") {
		str = str[2:]
		isWhileListRule = true
	}
	var matchHost string
	if strings.HasPrefix(str, "/") && strings.HasSuffix(str, "/") {
		str = str[1 : len(str)-1]
		rule = &regexRule{str}

	} else {
		if strings.HasPrefix(str, "||") {
			str = str[2:]
			if strings.HasSuffix(str, "/") {
				str = str[0 : len(str)-1]
			}
			if strings.Contains(str, "/") {
				log.Printf("Unsupported rule:%s", line)
				return nil
			}
			rule = &hostRule{str}
			matchHost = str
		} else if strings.HasPrefix(str, "|") || strings.HasSuffix(str, "|") {
			tmp := &urlWildcardRule{}
			if strings.HasPrefix(str, "|") {
				str = str[1:]
				tmp.prefixMatch = true
			}
			if strings.HasSuffix(str, "|") {
				str = str[0 : len(str)-1]
				tmp.suffixMatch = true
			}
			tmp.pattern = str
			tmp.init()
			u, err := url.Parse(str)
			if nil == err {
				matchHost = u.Host
			}
			rule = tmp
		} else {
			if !strings.Contains(str, "/") {
				rule = &hostRule{str}
				matchHost = str
			} else {
				tmp := &urlWildcardRule{pattern: str}
				tmp.init()
				rule = tmp
				matchHost = strings.SplitN(str, "/", 2)[0]
			}
		}
	}
	if isWhileListRule {
		rule = &whiteListRule{rule}
	}
	if len(matchHost) > 0 {
		if strings.Contains(matchHost, "*") {
			dr := domainRule{}
			dr.rule = rule
			dr.domain.pattern = matchHost
			gfw.domainPattenRules = append(gfw.domainPattenRules, dr)
			//log.Printf("###%s %s", str, matchHost)
		} else {
			if matchHost[0] == '.' {
				matchHost = matchHost[1:]
			}
			rs := gfw.fastMatchMap[matchHost]
			rs = append(rs, rule)
			gfw.fastMatchMap[matchHost] = rs
		}
	} else {
		//log.Printf("###%s", str)
		gfw.ruleList = append(gfw.ruleList, rule)
	}
	return nil
}

func (gfw *GFWList) loadContent(body []byte, base64Encoding bool) error {
	var err error
	if base64Encoding {
		body, err = base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return err
		}
	}

	reader := bufio.NewReader(strings.NewReader(string(body)))
	//i := 0
	for {
		line, _, err := reader.ReadLine()
		if nil != err {
			break
		}
		str := strings.TrimSpace(string(line))
		gfw.parse(str)
	}
	//log.Printf("####%d %d", len(gfw.fastMatchMap), len(gfw.ruleList))
	return nil
}

func (gfw *GFWList) Load(file string, base64Encoding bool) error {
	body, err := ioutil.ReadFile(file)
	if nil != err {
		return err
	}
	return gfw.loadContent(body, base64Encoding)
}
func (gfw *GFWList) LoadString(content string, base64Encoding bool) error {
	return gfw.loadContent([]byte(content), base64Encoding)
}

func (gfw *GFWList) Add(rule string) error {
	return gfw.parse(rule)
}

func (gfw *GFWList) getDomainRule(domain string) ([]gfwListRule, bool) {
	rules, exist := gfw.fastMatchMap[domain]
	if !exist {
		fs := strings.Split(domain, ".")
		if len(fs) > 2 {
			for i := 1; i < len(fs)-1; i++ {
				next := strings.Join(fs[i:], ".")
				rules, exist = gfw.fastMatchMap[next]
				if exist {
					break
				}
			}
		}
	}
	if !exist {
		for _, dr := range gfw.domainPattenRules {
			if dr.domain.matchDomain(domain) {
				return []gfwListRule{dr.rule}, true
			}
		}
	}
	return rules, exist
}

func (gfw *GFWList) IsDomainBlocked(domain string) bool {
	rs, exist := gfw.getDomainRule(domain)
	if exist {
		allWhilteList := true
		for _, r := range rs {
			if _, ok := r.(*whiteListRule); !ok {
				allWhilteList = false
			}
		}
		//log.Printf("### %s ==== %v", domain, allWhilteList)
		if allWhilteList {
			return false
		}
		return true
	}
	return false
}

func (gfw *GFWList) matchDomain(u *url.URL) (bool, bool) {
	domain := u.Host
	rules, exist := gfw.getDomainRule(domain)
	if exist {
		for _, rule := range rules {
			matched := rule.match(u)
			if matched {
				if _, ok := rule.(*whiteListRule); ok {
					return true, false
				}
				return true, true
			}
		}
	}
	return false, false
}

func (gfw *GFWList) Match(u *url.URL) (bool, bool) {
	matched, result := gfw.matchDomain(u)
	if matched {
		return matched, result
	}
	for _, rule := range gfw.ruleList {
		matched := rule.match(u)
		if matched {
			if _, ok := rule.(*whiteListRule); ok {
				return true, false
			}
			return true, true
		}
	}
	return false, false
}

func (gfw *GFWList) IsBlockedByGFW(req *http.Request) bool {
	gfw.mutex.Lock()
	defer gfw.mutex.Unlock()
	u := req.URL
	if len(u.Scheme) == 0 {
		u.Scheme = "http"
		if strings.EqualFold(req.Method, "CONNECT") {
			u.Scheme = "https"
		}
	}
	if len(u.Host) == 0 {
		u.Host = req.Host
	}
	_, result := gfw.Match(u)
	return result
}

func NewFromFile(file string, base64Encoding bool) (*GFWList, error) {
	gfw := &GFWList{}
	gfw.fastMatchMap = make(map[string][]gfwListRule)
	err := gfw.Load(file, base64Encoding)
	if nil != err {
		return nil, err
	}
	return gfw, nil
}
func NewFromString(str string, base64Encoding bool) (*GFWList, error) {
	gfw := &GFWList{}
	gfw.fastMatchMap = make(map[string][]gfwListRule)
	err := gfw.LoadString(str, base64Encoding)
	if nil != err {
		return nil, err
	}
	return gfw, nil
}
