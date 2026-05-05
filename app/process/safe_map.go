package process

import (
	"staploy-worker/app/proto"
	"sync"
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]*proto.Result
}

func (sm *SafeMap) Set(key string, taskProgression *proto.Result) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = taskProgression
}

func (sm *SafeMap) Get(key string) (*proto.Result, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	v, ok := sm.m[key]
	return v, ok
}
