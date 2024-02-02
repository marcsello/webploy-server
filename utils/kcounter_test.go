package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestKCounterMulti(t *testing.T) {

	kc := NewKCounter()

	closeChan := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			kc.Incr("a")
			defer kc.Dec("a")
			<-closeChan
		}()
	}

	time.Sleep(time.Millisecond * 2)

	for i := 10; i > 0; i-- {
		assert.Equal(t, uint(i), kc.Get("a"))
		closeChan <- false
		time.Sleep(time.Millisecond * 2)
	}

	assert.Equal(t, uint(0), kc.Get("a"))

}

func TestKCounterSimple(t *testing.T) {

	kc := NewKCounter()
	keys := []string{"a", "b", "c", "d"}

	for _, key := range keys {
		assert.Equal(t, uint(0), kc.Get(key))
		assert.Equal(t, uint(1), kc.Incr(key))
		assert.Equal(t, uint(1), kc.Get(key))
		assert.Equal(t, uint(2), kc.Incr(key))
		assert.Equal(t, uint(2), kc.Get(key))
	}

	keys = []string{"b", "d", "c", "a"}

	for _, key := range keys {
		assert.Equal(t, uint(2), kc.Get(key))
		assert.Equal(t, uint(1), kc.Dec(key))
		assert.Equal(t, uint(1), kc.Get(key))
		assert.Equal(t, uint(0), kc.Dec(key))
		assert.Equal(t, uint(0), kc.Get(key))
	}

}

func TestKCounterPanic(t *testing.T) {

	kc := NewKCounter()

	assert.Panics(t, func() {
		kc.Dec("a")
	})

	kc.Incr("b")
	kc.Dec("b")

	assert.Panics(t, func() {
		kc.Dec("b")
	})

}
