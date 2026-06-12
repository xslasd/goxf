package queue

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/log"
)

func TestMain(m *testing.M) {
	application.NewRuntime("test-app-id", "test-service", false, false, false, false, false)
	log.SetLogger(log.NewLogger())
	os.Exit(m.Run())
}

// TestCancelDelayedTask 测试延迟任务未执行时的取消
func TestCancelDelayedTask(t *testing.T) {
	memBroker := NewMemoryBroker()
	defer memBroker.Close()

	var executed int32

	worker, err := NewWorker(memBroker, func(ctx context.Context, payload string) error {
		atomic.StoreInt32(&executed, 1)
		return nil
	}, WithWorkerNum(1))
	if err != nil {
		t.Fatalf("failed to create worker: %v", err)
	}

	// 投递一个延迟 100ms 执行的任务
	taskID, err := worker.EnqueueAfter(context.Background(), "test-delay", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// 20ms 后进行取消
	time.Sleep(20 * time.Millisecond)
	err = worker.Cancel(context.Background(), taskID)
	if err != nil {
		t.Fatalf("failed to cancel: %v", err)
	}

	// 再等 150ms 观察是否被执行
	time.Sleep(150 * time.Millisecond)
	if atomic.LoadInt32(&executed) == 1 {
		t.Error("expected delayed task to be cancelled, but it was executed")
	}
}

// TestCancelExecutingTask 测试正在执行的任务的取消
func TestCancelExecutingTask(t *testing.T) {
	memBroker := NewMemoryBroker()
	defer memBroker.Close()

	started := make(chan struct{})
	cancelled := make(chan struct{})

	worker, err := NewWorker(memBroker, func(ctx context.Context, payload string) error {
		close(started) // 通知任务已开始执行
		select {
		case <-time.After(1 * time.Second):
			return nil
		case <-ctx.Done():
			close(cancelled) // 任务收到取消信号
			return ctx.Err()
		}
	}, WithWorkerNum(1))
	if err != nil {
		t.Fatalf("failed to create worker: %v", err)
	}

	taskID, err := worker.Enqueue(context.Background(), "test-exec")
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// 等待任务开始
	<-started

	// 取消该执行中的任务
	err = worker.Cancel(context.Background(), taskID)
	if err != nil {
		t.Fatalf("failed to cancel: %v", err)
	}

	// 等待任务被取消或超时
	select {
	case <-cancelled:
		// 成功捕获到了 Done 信号
	case <-time.After(500 * time.Millisecond):
		t.Error("expected task to be cancelled in time, but it timed out")
	}
}

// TestCancelDelayedTaskWithID 测试使用自定义 ID 延迟任务未执行时的取消
func TestCancelDelayedTaskWithID(t *testing.T) {
	memBroker := NewMemoryBroker()
	defer memBroker.Close()

	var executed int32

	worker, err := NewWorker(memBroker, func(ctx context.Context, payload string) error {
		atomic.StoreInt32(&executed, 1)
		return nil
	}, WithWorkerNum(1))
	if err != nil {
		t.Fatalf("failed to create worker: %v", err)
	}

	taskID := "custom-delay-id-123"
	err = worker.EnqueueAfterWithID(context.Background(), taskID, "test-delay", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	err = worker.Cancel(context.Background(), taskID)
	if err != nil {
		t.Fatalf("failed to cancel: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	if atomic.LoadInt32(&executed) == 1 {
		t.Error("expected delayed task to be cancelled, but it was executed")
	}
}

// TestCancelExecutingTaskWithID 测试使用自定义 ID 正在执行的任务的取消
func TestCancelExecutingTaskWithID(t *testing.T) {
	memBroker := NewMemoryBroker()
	defer memBroker.Close()

	started := make(chan struct{})
	cancelled := make(chan struct{})

	worker, err := NewWorker(memBroker, func(ctx context.Context, payload string) error {
		close(started)
		select {
		case <-time.After(1 * time.Second):
			return nil
		case <-ctx.Done():
			close(cancelled)
			return ctx.Err()
		}
	}, WithWorkerNum(1))
	if err != nil {
		t.Fatalf("failed to create worker: %v", err)
	}

	taskID := "custom-exec-id-456"
	err = worker.EnqueueWithID(context.Background(), taskID, "test-exec")
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	<-started

	err = worker.Cancel(context.Background(), taskID)
	if err != nil {
		t.Fatalf("failed to cancel: %v", err)
	}

	select {
	case <-cancelled:
	case <-time.After(500 * time.Millisecond):
		t.Error("expected task to be cancelled in time, but it timed out")
	}
}
