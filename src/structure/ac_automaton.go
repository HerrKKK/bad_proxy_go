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
	count := 0
	for i := range key {
		count += 1
		if curr.emit == true {
			return true
		}
		success, exist := curr.success[key[i]]
		for exist == false && curr != root {
			count += 1
			curr = curr.failure
			success, exist = curr.success[key[i]]
		}
		if exist == true {
			curr = success
		} // else: curr == root, curr = root
	}
	return false
}

func (root *ACAutomaton) build() {
	queue := NewQueue[*ACAutomaton](0, 1000)
	err := queue.Push(root)
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
				if s3 == root {
					s1.failure = root
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
