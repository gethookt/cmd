package check // import "hookt.dev/cmd/pkg/check"

import "sync"

type S struct {
	mu sync.Mutex

	Steps struct {
		OK   int
		Fail int
	}
}

func (s *S) OK() {
	s.mu.Lock()
	s.Steps.OK++
	s.mu.Unlock()
}

func (s *S) Fail() {
	s.mu.Lock()
	s.Steps.Fail++
	s.mu.Unlock()
}
