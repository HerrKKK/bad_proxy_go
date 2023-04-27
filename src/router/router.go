package router

import "strings"

type Matcher interface {
	Build() error
	MatchAny(key string) bool
}

type Trie struct {
	nodes   map[string]*Trie
	failure *Trie
	emit    bool
}

type DomainMatcher struct { // AC automaton
	root *Trie
}

func NewDomainMatcher() (matcher *DomainMatcher) {
	return &DomainMatcher{root: &Trie{nodes: make(map[string]*Trie)}}
}

func (matcher *DomainMatcher) MatchAny(host string) bool {
	tokens := strings.Split(host, ".")
	curr := matcher.root

	for i := len(tokens) - 1; i >= 0; i-- {
		next, exist := curr.nodes[tokens[i]]
		if exist == true {
			curr = next
		} else {
			curr = curr.failure
		}
		if curr == matcher.root {
			return false
		}
		if curr.emit == true {
			return true
		}
	}
	return false
}
