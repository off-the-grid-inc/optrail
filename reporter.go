package optrail

import (
	"sync"
	"sync/atomic"
)

// Reporter is a function that logs or further processes an OpTrail
type Reporter func(GenericMap)

type ReporterManager struct {
	sync.Mutex
	reporters atomic.Value
}

func newReporterManager() *ReporterManager {
	rm := &ReporterManager{}
	rm.reporters.Store(make([]Reporter, 0))
	return rm
}

func (rm *ReporterManager) RegisterReporter(reporter Reporter) {
	rm.Lock()
	defer rm.Unlock()

	old := rm.reporters.Load().([]Reporter)
	new := make([]Reporter, len(old)+1)
	for i, v := range old {
		new[i] = v
	}
	new[len(old)] = reporter

	rm.reporters.Store(new)
}

func (rm *ReporterManager) ClearReporters() {
	rm.Lock()
	rm.reporters.Store(make([]Reporter, 0))
	rm.Unlock()
}

func (rm *ReporterManager) getReporters() []Reporter {
	return rm.reporters.Load().([]Reporter)
}
