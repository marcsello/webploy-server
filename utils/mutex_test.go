package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

// Note that exact race conditions are very hard to test, so these tests only test for expected functionality

func TestKMutex_Lock_Simple(t *testing.T) {
	success := atomic.Bool{}
	km := NewKMutex()

	km.Lock("test")

	go func() {
		km.Lock("test")
		defer km.Unlock("test")
		success.Store(true)
	}()

	assert.False(t, success.Load())
	km.Unlock("test")

	time.Sleep(time.Millisecond * 30)
	assert.True(t, success.Load())

}

func TestKMutex_Lock_SimpleLock(t *testing.T) {
	success := atomic.Bool{}
	km := NewKMutex()

	km.Lock("test")

	go func() {
		km.Lock("test")
		defer km.Unlock("test")
		success.Store(true)
	}()

	assert.False(t, success.Load())
	time.Sleep(time.Millisecond * 30)
	assert.False(t, success.Load())

	km.Unlock("test")

	time.Sleep(time.Millisecond * 30)
	assert.True(t, success.Load())

}

func TestKMutex_Lock_SimpleOther(t *testing.T) {
	success := atomic.Bool{}
	km := NewKMutex()

	km.Lock("test")

	go func() {
		km.Lock("test2")
		defer km.Unlock("test2")
		success.Store(true)
	}()

	time.Sleep(time.Millisecond * 30)
	assert.True(t, success.Load())

}

func TestKMutex_Lock_ManySame(t *testing.T) {
	successCount := atomic.Int32{}
	km := NewKMutex()
	step := make(chan byte)
	proceeded := make(chan byte)

	for i := 0; i < 500; i++ {
		go func() {
			km.Lock("test")
			defer km.Unlock("test")
			<-step
			successCount.Add(1)
			proceeded <- 0
		}()
	}

	time.Sleep(time.Millisecond * 30)
	assert.Equal(t, int32(0), successCount.Load())
	for i := 0; i < 500; i++ {
		step <- 0
		<-proceeded
		assert.Equal(t, int32(i+1), successCount.Load())
	}

}

func TestKMutex_Lock_NotBlockingOther(t *testing.T) {
	oneContinued := atomic.Bool{}
	otherContinued := atomic.Bool{}
	km := NewKMutex()

	km.Lock("one")

	go func() {
		km.Lock("one")
		defer km.Unlock("one")
		oneContinued.Store(true)
	}()
	go func() {
		km.Lock("other")
		defer km.Unlock("other")
		otherContinued.Store(true)
	}()

	time.Sleep(time.Millisecond * 10)
	assert.False(t, oneContinued.Load())
	assert.True(t, otherContinued.Load())

	km.Unlock("one")
	time.Sleep(time.Millisecond * 10)
	assert.True(t, oneContinued.Load())
	assert.True(t, otherContinued.Load())

}

func TestKMutex_Lock_ManySome(t *testing.T) {
	successCount := atomic.Int32{}
	km := NewKMutex()
	proceeded := make(chan int)

	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("%d", i)
		km.Lock(key)
	}

	for i := 0; i < 500; i++ {
		go func(id int) {
			key := fmt.Sprintf("%d", id)
			km.Lock(key)
			defer km.Unlock(key)
			successCount.Add(1)
			proceeded <- id
		}(i)
	}

	time.Sleep(time.Millisecond * 30)
	assert.Equal(t, int32(0), successCount.Load())
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("%d", i)
		km.Unlock(key)
		id := <-proceeded
		assert.Equal(t, i, id)
		assert.Equal(t, int32(i+1), successCount.Load())
	}

}

func TestKMutex_Unlock_Panic(t *testing.T) {
	km := NewKMutex()

	assert.Panics(t, func() {
		km.Unlock("a")
	})

	km.Lock("b")
	km.Unlock("b")

	assert.Panics(t, func() {
		km.Unlock("b")
	})
}
