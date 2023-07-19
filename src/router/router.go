package router

import (
	"go_proxy/structure"
	"log"
	"regexp"
	"strings"
)

type Config struct {
	Tag   string   `json:"tag"`
	Rules []string `json:"rules"`
}

type Matcher interface {
	MatchAny(key string) bool
}

type FullMatcher map[string]bool

func (matcher FullMatcher) MatchAny(key string) bool {
	_, exist := matcher[key]
	return exist
}

func NewFullMatcher(fullDomains []string) Matcher {
	var fullMatcher FullMatcher = make(map[string]bool, len(fullDomains))
	for _, domain := range fullDomains {
		fullMatcher[domain] = true
	}
	return fullMatcher
}

type RegexMatcher []*regexp.Regexp

func NewRegexMatcher(RegexStrs []string) Matcher {
	var regexMatcher RegexMatcher = make([]*regexp.Regexp, len(RegexStrs))
	for i, ex := range RegexStrs {
		regexMatcher[i] = regexp.MustCompile(ex)
	}
	return regexMatcher
}

func (matcher RegexMatcher) MatchAny(key string) bool {
	for _, re := range matcher {
		if re.MatchString(key) {
			return true
		}
	}
	return false
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

func NewRouter(tag string, ruleNames []string, routerPath string) (router *Router, err error) {
	//allRules, err := readAllFromFile("rules")
	allRules, err := readAllFromGob(routerPath)
	if err != nil {
		log.Println("failed to read rules from file", err)
		return
	}

	fullDomains := make([]string, 0)
	domains := make([]string, 0)
	RegexStrs := make([]string, 0)
	for _, ruleName := range ruleNames {
		ruleList, _ := allRules[strings.ToUpper(ruleName)]
		if ruleList == nil {
			continue
		}
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
	router.matchers = append(router.matchers, NewFullMatcher(fullDomains))
	router.matchers = append(router.matchers, NewRegexMatcher(RegexStrs))
	router.matchers = append(router.matchers, structure.NewACAutomaton(domains))
	return
}
