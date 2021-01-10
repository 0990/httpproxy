package httpproxy

import (
	"strings"
)

type Matcher interface {
	Match(s string) bool
}

type domainMatcher string

func (m domainMatcher) Match(s string) bool {
	pattern := string(m)
	if !strings.HasSuffix(s, pattern) {
		return false
	}
	return len(s) == len(pattern) || s[len(s)-len(pattern)-1] == '.'
}

type matcherManger struct {
	matchAny bool
	matches  []Matcher
}

func newMatcherManager(hosts []string) *matcherManger {
	mm := &matcherManger{}

	for _, v := range hosts {
		if v == "*" {
			mm.matchAny = true
		} else {
			mm.matches = append(mm.matches, domainMatcher(v))
		}
	}
	return mm
}

func (mm *matcherManger) Match(host string) bool {
	if mm.matchAny {
		return true
	}

	for _, v := range mm.matches {
		if v.Match(host) {
			return true
		}
	}
	return false
}
