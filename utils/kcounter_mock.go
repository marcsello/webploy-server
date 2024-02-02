package utils

type KCounterMock struct {
	IncrFn func(string) uint
	DecFn  func(string) uint
	GetFn  func(string) uint
}

func (kc *KCounterMock) Incr(key string) uint {
	if kc.IncrFn != nil {
		return kc.IncrFn(key)
	}
	return 0
}

func (kc *KCounterMock) Dec(key string) uint {
	if kc.DecFn != nil {
		return kc.DecFn(key)
	}
	return 0
}

func (kc *KCounterMock) Get(key string) uint {
	if kc.GetFn != nil {
		return kc.GetFn(key)
	}
	return 0
}
