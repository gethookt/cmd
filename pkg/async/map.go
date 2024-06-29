package async

import (
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return sync.NewCond(&sync.Mutex{})
	},
}

type Map struct {
	m sync.Map
}

func (m *Map) Store(key, value any) {
	prev, _ := m.m.Swap(key, value)
	cond, ok := prev.(*sync.Cond)
	if !ok {
		return
	}

	cond.L.Lock()
	cond.Broadcast()
	cond.L.Unlock()

	pool.Put(cond)
}

func (m *Map) Load(key any) (any, bool) {
	value, _ := m.m.LoadOrStore(key, pool.Get())
	cond, wait := value.(*sync.Cond)
	if !wait {
		return value, true
	}

	cond.L.Lock()
	for {
		v, _ := m.m.Load(key)
		if _, ok := v.(*sync.Cond); !ok {
			break
		}

		cond.Wait()
	}
	cond.L.Unlock()

	return m.m.Load(key)
}
