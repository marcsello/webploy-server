package utils

import "sync"

type KCounter interface {
	Incr(key string) uint
	Dec(key string) uint
	Get(key string) uint
}

type KCounterImpl struct {
	m    *sync.RWMutex // I usually prefer atomic solutions, but I could not solve this that way
	vals map[string]uint
}

func NewKCounter() KCounter {
	return &KCounterImpl{
		m:    &sync.RWMutex{},
		vals: make(map[string]uint),
	}
}

func (kc *KCounterImpl) Incr(key string) uint {
	kc.m.Lock()
	defer kc.m.Unlock()
	val, ok := kc.vals[key]
	if ok {
		val++
	} else {
		val = 1
	}
	kc.vals[key] = val
	return val
}

func (kc *KCounterImpl) Dec(key string) uint {
	kc.m.Lock()
	defer kc.m.Unlock()
	val, ok := kc.vals[key]
	if ok {
		val--
	} else { // should not happen if used properly
		panic("decrementing of zero key")
	}
	if val == 0 {
		delete(kc.vals, key)
	} else {
		kc.vals[key] = val
	}
	return val
}

func (kc *KCounterImpl) Get(key string) uint {
	kc.m.RLock()
	defer kc.m.RUnlock()
	val, ok := kc.vals[key]
	if ok {
		return val
	} else {
		return 0
	}
}
