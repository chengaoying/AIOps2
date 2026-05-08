package queue

import (
	"sync"
	"testing"
	"time"

	"aiops2/collector/internal/model"
)

func TestMemoryQueue_New(t *testing.T) {
	q := New(100)
	if q.Len() != 0 {
		t.Errorf("New queue len = %d, want 0", q.Len())
	}
	if q.Cap() != 100 {
		t.Errorf("New queue cap = %d, want 100", q.Cap())
	}
}

func TestMemoryQueue_Push(t *testing.T) {
	q := New(3)

	q.Push(&model.JobMeta{JobID: "job1"})
	if q.Len() != 1 {
		t.Errorf("Push len = %d, want 1", q.Len())
	}

	q.Push(&model.JobMeta{JobID: "job2"})
	if q.Len() != 2 {
		t.Errorf("Push len = %d, want 2", q.Len())
	}
}

func TestMemoryQueue_Pop(t *testing.T) {
	q := New(10)
	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})

	job := q.Pop()
	if job.JobID != "job1" {
		t.Errorf("Pop job = %v, want job1", job.JobID)
	}
	if q.Len() != 1 {
		t.Errorf("Pop len = %d, want 1", q.Len())
	}
}

func TestMemoryQueue_FIFO(t *testing.T) {
	q := New(100)

	for i := 1; i <= 5; i++ {
		q.Push(&model.JobMeta{JobID: "job" + string(rune('0'+i))})
	}

	for i := 1; i <= 5; i++ {
		job := q.Pop()
		expected := "job" + string(rune('0'+i))
		if job.JobID != expected {
			t.Errorf("FIFO: got %s, want %s", job.JobID, expected)
		}
	}
}

func TestMemoryQueue_Overflow(t *testing.T) {
	q := New(3)

	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})
	q.Push(&model.JobMeta{JobID: "job3"})
	q.Push(&model.JobMeta{JobID: "job4"})

	if q.Len() != 3 {
		t.Errorf("Overflow len = %d, want 3", q.Len())
	}

	job := q.Pop()
	if job.JobID != "job2" {
		t.Errorf("Overflow pop = %s, want job2 (oldest dropped)", job.JobID)
	}
}

func TestMemoryQueue_PopEmpty(t *testing.T) {
	q := New(10)
	job := q.Pop()
	if job != nil {
		t.Errorf("Pop empty queue = %v, want nil", job)
	}
}

func TestMemoryQueue_Full(t *testing.T) {
	q := New(2)
	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})

	if !q.Full() {
		t.Error("Queue should be full")
	}

	q.Push(&model.JobMeta{JobID: "job3"})
	if !q.Full() {
		t.Error("Queue should still be full after overflow")
	}
}

func TestMemoryQueue_Empty(t *testing.T) {
	q := New(10)
	if !q.Empty() {
		t.Error("New queue should be empty")
	}

	q.Push(&model.JobMeta{JobID: "job1"})
	if q.Empty() {
		t.Error("Queue with item should not be empty")
	}
}

func TestMemoryQueue_Concurrent(t *testing.T) {
	q := New(1000)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				q.Push(&model.JobMeta{JobID: "job"})
			}
		}()
	}

	wg.Wait()

	if q.Len() != 1000 {
		t.Errorf("Concurrent push len = %d, want 1000", q.Len())
	}
}

func TestMemoryQueue_ConcurrentPop(t *testing.T) {
	q := New(1000)
	for i := 0; i < 500; i++ {
		q.Push(&model.JobMeta{JobID: "job"})
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				q.Pop()
			}
		}()
	}

	wg.Wait()

	if q.Len() != 0 {
		t.Errorf("Concurrent pop len = %d, want 0", q.Len())
	}
}

func TestMemoryQueue_Peek(t *testing.T) {
	q := New(10)
	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})

	job := q.Peek()
	if job.JobID != "job1" {
		t.Errorf("Peek = %s, want job1", job.JobID)
	}
	if q.Len() != 2 {
		t.Errorf("Peek should not remove item, len = %d", q.Len())
	}
}

func TestMemoryQueue_Clear(t *testing.T) {
	q := New(10)
	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})

	q.Clear()
	if q.Len() != 0 {
		t.Errorf("Clear len = %d, want 0", q.Len())
	}
	if !q.Empty() {
		t.Error("Clear queue should be empty")
	}
}

func TestMemoryQueue_BatchPush(t *testing.T) {
	q := New(100)
	jobs := make([]*model.JobMeta, 50)
	for i := 0; i < 50; i++ {
		jobs[i] = &model.JobMeta{JobID: "job" + string(rune('0'+i))}
	}

	q.BatchPush(jobs)
	if q.Len() != 50 {
		t.Errorf("BatchPush len = %d, want 50", q.Len())
	}
}

func TestMemoryQueue_BatchPop(t *testing.T) {
	q := New(100)
	for i := 0; i < 50; i++ {
		q.Push(&model.JobMeta{JobID: "job" + string(rune('0'+i))})
	}

	jobs := q.BatchPop(10)
	if len(jobs) != 10 {
		t.Errorf("BatchPop count = %d, want 10", len(jobs))
	}
	if q.Len() != 40 {
		t.Errorf("BatchPop remaining len = %d, want 40", q.Len())
	}
}

func TestMemoryQueue_BatchPopMoreThanAvailable(t *testing.T) {
	q := New(10)
	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})

	jobs := q.BatchPop(100)
	if len(jobs) != 2 {
		t.Errorf("BatchPop more than available = %d, want 2", len(jobs))
	}
}

func TestMemoryQueue_Stats(t *testing.T) {
	q := New(10)
	q.Push(&model.JobMeta{JobID: "job1"})
	q.Push(&model.JobMeta{JobID: "job2"})

	stats := q.Stats()
	if stats["len"] != 2 {
		t.Errorf("Stats len = %v, want 2", stats["len"])
	}
	if stats["cap"] != 10 {
		t.Errorf("Stats cap = %v, want 10", stats["cap"])
	}
}

func TestMemoryQueue_EvictionCount(t *testing.T) {
	q := New(3)

	for i := 0; i < 10; i++ {
		q.Push(&model.JobMeta{JobID: "job"})
	}

	stats := q.Stats()
	if stats["evictions"].(int) != 7 {
		t.Errorf("Evictions = %d, want 7", stats["evictions"])
	}
}

func TestMemoryQueue_PushWithTimestamp(t *testing.T) {
	q := New(10)
	start := time.Now()
	q.Push(&model.JobMeta{JobID: "job1"})
	end := time.Now()

	item := q.Peek()
	if item == nil {
		t.Error("Peek should return item")
	}
}

func TestMemoryQueue_RingBufferWrap(t *testing.T) {
	q := New(3)

	for i := 0; i < 100; i++ {
		q.Push(&model.JobMeta{JobID: string(rune('a' + i%26))})
	}

	stats := q.Stats()
	if stats["evictions"].(int) != 97 {
		t.Errorf("Ring buffer wrap evictions = %d, want 97", stats["evictions"])
	}
}
