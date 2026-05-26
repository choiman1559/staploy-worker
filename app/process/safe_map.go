package process

import (
	"staploy-worker/app/proto"
	"sync"
)

type SafeQueue struct {
	mu    sync.RWMutex
	m     map[string]*proto.Result
	order []string
	max   int
}

func NewSafeQueue() *SafeQueue {
	return &SafeQueue{
		m:     make(map[string]*proto.Result),
		order: make([]string, 0, 64),
		max:   64,
	}
}

func (sq *SafeQueue) Enqueue(key string, taskProgression *proto.Result) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if _, exists := sq.m[key]; exists {
		for i, k := range sq.order {
			if k == key {
				sq.order = append(sq.order[:i], sq.order[i+1:]...)
				break
			}
		}
	}

	if len(sq.order) >= sq.max {
		oldestKey := sq.order[0]
		sq.order = sq.order[1:]
		delete(sq.m, oldestKey)
	}

	sq.m[key] = taskProgression
	sq.order = append(sq.order, key)
}

func (sq *SafeQueue) Dequeue() (string, *proto.Result, bool) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if len(sq.order) == 0 {
		return "", nil, false
	}

	key := sq.order[0]
	sq.order = sq.order[1:]

	v := sq.m[key]
	delete(sq.m, key)

	return key, v, true
}

func (sq *SafeQueue) Get(key string) (*proto.Result, bool) {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	v, ok := sq.m[key]
	return v, ok
}

func (sq *SafeQueue) Len() int {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return len(sq.order)
}
