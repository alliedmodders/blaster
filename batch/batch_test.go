package batch

import (
	"sync"
	"testing"
	"time"
)

type MyBatch struct {
}

func (this *MyBatch) Len() int {
	return 10
}
func (this *MyBatch) Item(index int) interface{} {
	return index + 1
}

func TestBasic(t *testing.T) {
	var lock sync.Mutex
	items := make([]interface{}, 0)

	bp := NewBatchProcessor(func(item interface{}) {
		lock.Lock()
		defer lock.Unlock()

		items = append(items, item)
	}, 10)

	bp.AddBatch(&MyBatch{})
	bp.AddBatch(&MyBatch{})
	bp.AddBatch(&MyBatch{})
	bp.AddBatch(&MyBatch{})
	bp.Finish()

	if len(items) != 40 {
		t.Errorf("expected 40 items, got %d", len(items))
	}

	count := 0
	for _, item := range items {
		if item.(int) == 5 {
			count++
		}
	}
	if count != 4 {
		t.Errorf("expected to see item #5 four times, but got it %d time(s)", count)
	}
}

func TestTerminate(t *testing.T) {
	bp := NewBatchProcessor(func(item interface{}) {
		time.Sleep(time.Second)
	}, 10)

	bp.AddBatch(&MyBatch{})
	bp.AddBatch(&MyBatch{})
	bp.AddBatch(&MyBatch{})
	bp.AddBatch(&MyBatch{})

	// We should not block here. If we do, the test will be extremely slow.
	bp.Terminate()
}
