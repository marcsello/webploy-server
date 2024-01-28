package jobs

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"sync/atomic"
	"testing"
	"time"
)

type testJob struct {
	invoked     bool
	invokeCount atomic.Int32
	delay       time.Duration
}

func (j *testJob) Run(logger *zap.Logger) {
	j.invoked = true
	j.invokeCount.Add(1)
	logger.Info("ok")
	if j.delay > 0 {
		time.Sleep(j.delay)
	}
}

func TestJobWrapper_Happy(t *testing.T) {

	testLogger := zaptest.NewLogger(t)

	tj := &testJob{}

	wrappedJob := wrapJob(testLogger, tj)

	assert.Equal(t, "testJob", wrappedJob.name)
	assert.False(t, tj.invoked)
	assert.Equal(t, int32(0), tj.invokeCount.Load())
	assert.Equal(t, uint64(0), wrappedJob.execId.Load())

	wrappedJob.Run()

	assert.True(t, tj.invoked)
	assert.Equal(t, int32(1), tj.invokeCount.Load())
	assert.Equal(t, uint64(1), wrappedJob.execId.Load())

	wrappedJob.Run()

	assert.True(t, tj.invoked)
	assert.Equal(t, int32(2), tj.invokeCount.Load())
	assert.Equal(t, uint64(2), wrappedJob.execId.Load())

}

func TestJobWrapper_Overrun(t *testing.T) {

	testLogger := zaptest.NewLogger(t)

	tj := &testJob{
		delay: time.Second * 1,
	}

	wrappedJob := wrapJob(testLogger, tj)

	assert.Equal(t, "testJob", wrappedJob.name)
	assert.False(t, tj.invoked)
	assert.Equal(t, uint64(0), wrappedJob.execId.Load())
	assert.Equal(t, int32(0), tj.invokeCount.Load())

	go wrappedJob.Run()
	go wrappedJob.Run()

	time.Sleep(time.Second * 2)

	assert.True(t, tj.invoked)
	assert.Equal(t, uint64(2), wrappedJob.execId.Load())
	assert.Equal(t, int32(1), tj.invokeCount.Load())

}
