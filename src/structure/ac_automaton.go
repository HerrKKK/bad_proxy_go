package structure

func NewTrie() *Trie {
	trie := &Trie{
		success: make(map[uint8]*Trie),
		failure: nil,
		emit:    false,
	}
	trie.failure = trie
	return trie
}

type Trie struct {
	success map[uint8]*Trie
	failure *Trie
	emit    bool
}

func (trie *Trie) Add(value string) {
	curr := trie
	for i := range value {
		ch := value[i]
		_, exist := curr.success[ch]
		if exist == false {
			curr.success[ch] = NewTrie()
		}
		curr = curr.success[ch]
		if i == len(value)-1 {
			curr.emit = true
		}
	}
}

type ACAutomaton struct {
	root *Trie
}

func (ac *ACAutomaton) MatchAny(key string) bool {
	curr := ac.root
	count := 0
	for i := range key {
		count += 1
		if curr.emit == true {
			return true
		}
		success, exist := curr.success[key[i]]
		for exist == false && curr != ac.root {
			count += 1
			curr = curr.failure
			success, exist = curr.success[key[i]]
		}
		if exist == true {
			curr = success
		} // else: curr == ac.root, curr = ac.root
	}
	return false
}

func NewACAutomaton(patterns []string) *ACAutomaton {
	ac := &ACAutomaton{root: NewTrie()}
	for _, pattern := range patterns {
		ac.root.Add(pattern)
	}
	ac.Build()
	return ac
}

func (ac *ACAutomaton) Build() {
	queue := NewQueue[*Trie](0, 1000)
	err := queue.Push(ac.root)
	if err != nil {
		panic(err) // should never happen
	}
	for queue.Size() != 0 { // bfs without layer
		s2 := queue.Pop()               // s2: s1 is s2's success state with c
		for c, s1 := range s2.success { // s1: state we are finding failure for
			err = queue.Push(s1)
			if err != nil {
				panic(err)
			}
			s3 := s2.failure // s3: s2's failure state
			for {
				if s3 == ac.root {
					s1.failure = ac.root
					break
				}
				s4, exist := s3.success[c] // s4: s3's success state with c
				if exist == true {
					s1.failure = s4
					break
				}
				s3 = s3.failure
			}
		}
	}
}
