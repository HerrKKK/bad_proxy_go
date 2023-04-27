package router

import (
	"go_proxy/structure"
	"log"
	"regexp"
	"strings"
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

type Router struct {
	Tag      string
	matchers []Matcher
}

func (router Router) MatchAny(key string) bool {
	for _, m := range router.matchers {
		if m.MatchAny(key) {
			return true
		}
	}
	return false
}

func NewRouter(tag string, ruleNames []string) (router *Router, err error) {
	allRules, err := readAllFromFile("rules")
	if err != nil {
		log.Println("failed to read rules from file")
		return
	}

	fullDomains := make([]string, 0)
	domains := make([]string, 0)
	RegexStrs := make([]string, 0)
	for _, ruleName := range ruleNames {
		ruleList, _ := allRules[strings.ToUpper(ruleName)]
		for _, entry := range ruleList.Entry {
			switch entry.Type {
			case "full":
				fullDomains = append(fullDomains, entry.Value)
			case "domain":
				domains = append(domains, entry.Value)
			case "regexp":
				RegexStrs = append(RegexStrs, entry.Value)
			default:
				// include here
			}
		}
	}
	router = &Router{
		Tag:      tag,
		matchers: make([]Matcher, 0),
	}
	log.Printf(
		"%d rules with %d domains loaded\n",
		len(allRules),
		len(fullDomains)+len(RegexStrs)+len(domains),
	)
	router.matchers = append(router.matchers, *NewFullMatcher(fullDomains))
	router.matchers = append(router.matchers, *NewRegexMatcher(RegexStrs))
	//router.matchers = append(router.matchers, *NewACAutomatonMatcher(domains))
	return
}
