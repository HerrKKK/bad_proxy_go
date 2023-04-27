package structure

func NewTrie() *ACAutomaton {
	trie := &ACAutomaton{
		success: make(map[uint8]*ACAutomaton),
		failure: nil,
		emit:    false,
	}
	trie.failure = trie
	return trie
}

type ACAutomaton struct {
	success map[uint8]*ACAutomaton
	failure *ACAutomaton
	emit    bool
}

func (root *ACAutomaton) Add(value string) {
	curr := root
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

func NewACAutomaton(patterns []string) *ACAutomaton {
	ac := NewTrie()
	for _, pattern := range patterns {
		ac.Add(pattern)
	}
	ac.build()
	return ac
}

func (root *ACAutomaton) MatchAny(key string) bool {
	curr := root
	for i := range key {
		success, exist := curr.success[key[i]]
		for exist == false && curr != root {
			curr = curr.failure
			success, exist = curr.success[key[i]]
		}
		if exist == true {
			curr = success
		} // else: curr == root, curr = root
		if curr.emit == true {
			return true
		}
	}
	return false
}

func (root *ACAutomaton) build() {
	queue := NewQueue[*ACAutomaton](0, 1000)
	err := queue.Push(root)
	if err != nil {
		panic(err)
	}
	for queue.Size() != 0 { // bfs without layer
		s2 := queue.Pop()
		if s2 == nil {
			panic("s2 is nil")
		}
		for c, s1 := range s2.success {
			err = queue.Push(s1)
			if err != nil {
				panic(err)
			}
			s3 := s2.failure
			for {
				if s3 == root {
					s1.failure = root
					break
				}
				s4, exist := s3.success[c]
				if exist == true {
					s1.failure = s4
					break
				}
				s3 = s3.failure
			}
		}
	}
}
