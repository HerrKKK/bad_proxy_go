package proxy

type RuleConfig struct {
	Tag   string   `json:"tag"`
	Rules []string `json:"rules"`
}

type Rule struct {
	Tag     string
	Domains map[string]bool
}

func NewRule(tag string, ruleName []string) (rule Rule) {
	rule = Rule{Tag: tag, Domains: make(map[string]bool)}
	for _, name := range ruleName {
		rule.Domains[name] = true
	}
	return
}

type Router struct {
	rules []Rule
}

func NewRouter(configs []RuleConfig) (router Router) {
	router = Router{rules: make([]Rule, 0)}
	for _, config := range configs {
		router.rules = append(router.rules, NewRule(config.Tag, config.Rules))
	}
	return
}

func (router Router) route(domain string) (tag string) {
	for _, rule := range router.rules {
		_, exist := rule.Domains[domain]
		if exist == true {
			return rule.Tag
		}
	}
	return ""
}
