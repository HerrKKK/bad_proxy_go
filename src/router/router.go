package router

import (
	"go_proxy/structure"
	"regexp"
)

type Matcher interface {
	MatchAny(key string) bool
}

type FullMatcher map[string]bool

func (matcher *FullMatcher) MatchAny(key string) bool {
	_, exist := (*matcher)[key]
	return exist
}

func NewFullMatcher(fullDomains []string) *Matcher {
	var matcher Matcher
	var fullMatcher FullMatcher
	m := make(map[string]bool, len(fullDomains))
	fullMatcher = m
	for _, domain := range fullDomains {
		fullMatcher[domain] = true
	}
	matcher = &fullMatcher
	return &matcher
}

type RegexMatcher []*regexp.Regexp

func NewRegexMatcher(RegexStrs []string) *Matcher {
	var matcher Matcher
	var regexMatcher RegexMatcher
	regexMatcher = make([]*regexp.Regexp, len(RegexStrs))
	for i, ex := range RegexStrs {
		regexMatcher[i] = regexp.MustCompile(ex)
	}
	matcher = &regexMatcher
	return &matcher
}

func (matcher *RegexMatcher) MatchAny(key string) bool {
	for _, re := range *matcher {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}

func NewACAutomatonMatcher(domains []string) *Matcher {
	var matcher Matcher
	var acMatcher structure.ACAutomaton
	acMatcher = *structure.NewACAutomaton(domains)
	matcher = &acMatcher
	return &matcher
}

type Router []Matcher

func (router Router) MatchAny(key string) bool {
	for _, m := range router {
		if m.MatchAny(key) {
			return true
		}
	}
	return false
}

func NewRouter(
	fullDomains []string,
	RegexStrs []string,
	domains []string,
) *Router {
	matchers := make([]Matcher, 0)
	matchers = append(matchers, *NewFullMatcher(fullDomains))
	matchers = append(matchers, *NewRegexMatcher(RegexStrs))
	matchers = append(matchers, *NewACAutomatonMatcher(domains))
	return (*Router)(&matchers)
}
