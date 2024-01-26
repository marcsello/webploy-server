package utils

import (
	"sync"
	"time"
)

// This is my own lame implementation of KMutex, because I couldn't find any better that
// - supports arbitrary number of locks (we only store as many as the number of our deployments)
// - does not have flawed lock logic
//    - many implementations forget to unlock the lock they use to protect the map, while waiting for a key, essentially creating a global lock
//    - some other implementations aren't safe from race conditions, because they put unlocked keys in shared maps, which could be stolen by other routines
// This implementation might not be as efficient, but at least fixes the problems I've found in others...

type KMutex struct {
	m     *sync.Mutex
	locks map[string]*sync.Mutex
}

func NewKMutex() *KMutex {
	return &KMutex{
		m:     &sync.Mutex{},
		locks: make(map[string]*sync.Mutex),
	}
}

func (km *KMutex) Lock(key string) {
	km.m.Lock()

	l, ok := km.locks[key]
	if !ok {
		l = &sync.Mutex{}
		l.Lock()
		km.locks[key] = l
		km.m.Unlock()
		return
	}
	go func() {
		// we have to make sure we only unlock after the lock is called
		// So it is not possible to accidentally lock a lock that just been unlocked and lost the reference to it
		time.Sleep(time.Millisecond * 2)
		km.m.Unlock()
	}()
	l.Lock() // we only store locked keys in the map which are unlocked only when they are removed from the map, so this lock must be locked already
	l.Unlock()

	// after the lock has been unlocked, retry locking it
	km.Lock(key)
}

func (km *KMutex) Unlock(key string) {
	km.m.Lock()
	defer km.m.Unlock()

	l := km.locks[key]
	delete(km.locks, key)
	l.Unlock()
}
