package optrail

import (
	"fmt"
	"sync"
)

type goId = uint64

type trailManager struct {
	sync.RWMutex
	trails  map[goId]*opTrail
	globals GenericMap
}

func newTrailManager() *trailManager {
	return &trailManager{
		trails:  make(map[goId]*opTrail),
		globals: make(GenericMap),
	}
}

func (m *trailManager) beginOpTrail() *opTrail {
	id := curGoroutineID()
	m.Lock()
	defer m.Unlock()
	c := m.trails[id]
	if c == nil {
		c = makeOpTrail(id, nil, nil)
		m.trails[id] = c
	}
	return c
}

func (m *trailManager) removeTrail(id goId) {
	m.Lock()
	delete(m.trails, id)
	m.Unlock()
}

func (m *trailManager) getTrail(id goId) *opTrail {
	m.RLock()
	c := m.trails[id]
	m.RUnlock()
	return c
}

func (m *trailManager) setTrail(id goId, t *opTrail) {
	m.Lock()
	m.trails[id] = t
	m.Unlock()
}

func (m *trailManager) putGlobal(key string, value interface{}) {
	m.Lock()
	m.globals[key] = value
	m.Unlock()
}

func (m *trailManager) clearGlobals() {
	m.Lock()
	m.trails = make(map[goId]*opTrail)
	m.Unlock()
}

func (m *trailManager) killAll() {
	m.Lock()
	m.trails = make(map[goId]*opTrail)
	m.Unlock()
}

func (m *trailManager) printTrails() {
	m.RLock()
	if len(m.trails) > 0 {
		for k := range m.trails {
			fmt.Printf("Trail ID %v\n", k)
		}
	} else {
		fmt.Println("No trails.")
	}
	m.RUnlock()
}
