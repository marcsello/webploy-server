package utils

// source: https://medium.com/@petrlozhkin/kmutex-lock-mutex-by-unique-id-408467659c24
// Sadly, it is not available from the author's GitHub repo, so we have to copy

import (
	"sync"
)

type Kmutex struct {
	locks *sync.Map
}

func NewKmutex() Kmutex {
	m := sync.Map{}
	return Kmutex{&m}
}

func (s *Kmutex) Unlock(key interface{}) {
	l, exist := s.locks.Load(key)
	if !exist {
		panic("kmutex: unlock of unlocked mutex")
	}
	l_ := l.(*sync.Mutex)
	s.locks.Delete(key)
	l_.Unlock()
}

func (s *Kmutex) Lock(key interface{}) {
	m := sync.Mutex{}
	m_, _ := s.locks.LoadOrStore(key, &m)
	mm := m_.(*sync.Mutex)
	mm.Lock()
	if mm != &m {
		mm.Unlock()
		s.Lock(key)
		return
	}
	return
}
