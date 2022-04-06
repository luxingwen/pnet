package common

import "sync"

type LifeCycle struct {
	StartOnce sync.Once
	StopOnce  sync.Once
	stopped   bool
	stopLock  sync.RWMutex
	ready     bool
	readyLock sync.RWMutex
}

func (l *LifeCycle) Stop() {
	l.stopLock.Lock()
	l.stopped = true
	l.stopLock.Unlock()
}

func (l *LifeCycle) IsStopped() bool {
	l.stopLock.RLock()
	defer l.stopLock.RUnlock()
	return l.stopped
}

func (l *LifeCycle) SetReady(ready bool) {
	l.readyLock.Lock()
	l.ready = ready
	l.readyLock.Unlock()
}

func (l *LifeCycle) IsReady() bool {
	l.readyLock.RLock()
	defer l.readyLock.RUnlock()
	return l.ready
}
