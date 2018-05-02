package optrail

import (
	"sync"
)

var (
	tManager = newTrailManager()
	rManager = newReporterManager()
)

// Begin creates a new OpTrail
func Begin(name string) OpTrail {
	return tManager.BeginOpTrail().Here("name", name)
}

func PutGlobal(key string, value interface{}) {
	tManager.putGlobal(key, value)
}

func CurrentTrail() *opTrail {
	return tManager.getTrail(curGoroutineID())
}

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

func (m *trailManager) BeginOpTrail() *opTrail {
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

// OpTrail is an operation that can cross API boundaries, representation formats, storage
// and tranmission, and remembers its history
type OpTrail interface {
	// Succeed marks this OpTrail as succeeded
	Succeed()

	// FailIf marks this OpTrail as failed if the given err is not nil. If FailIf is
	// called multiple times, only the first one will be effective.
	FailIf(err error) error

	// Here marks this point in time and adds a specific key value to provide
	// information about it
	Here(key string, value interface{}) OpTrail

	// Vanish makes the trail disappear, no more actions can be done with it
	Vanish()

	// Go starts a given thunk on a new goroutine.
	Go(fn func())
}

type opTrail struct {
	sync.RWMutex

	id           goId
	parent       *opTrail
	branchedFrom *opTrail
	data         TimestampedMap
}

func makeOpTrail(id uint64, parent *opTrail, branchedFrom *opTrail) *opTrail {
	return &opTrail{
		id:           id,
		parent:       parent,
		branchedFrom: branchedFrom,
		data:         make(TimestampedMap),
	}
}

func (t *opTrail) Succeed() {
	t.Lock()
	t.data["succeeded"] = newTimestamped(true)
	t.Unlock()
	t.finalize()
	return
}

func (t *opTrail) FailIf(err error) error {
	t.Lock()
	vals := newTimestampedMulti(false, err)
	t.data["succeeded"] = vals[0]
	t.data["error"] = vals[1]
	t.Unlock()
	t.finalize()
	return err
}

func (t *opTrail) Here(key string, value interface{}) OpTrail {
	t.Lock()
	t.data[key] = newTimestamped(value)
	t.Unlock()
	return t
}

func (t *opTrail) Go(fn func()) {
	go func() {
		id := curGoroutineID()
		next := makeOpTrail(id, nil, t)
		tManager.Lock()
		tManager.trails[id] = next
		tManager.Unlock()
		fn()
		tManager.removeTrail(id)
	}()
}

func (t *opTrail) Vanish() {
	tManager.removeTrail(t.id)
}

func (t *opTrail) finalize() {
	reporters := rManager.getReporters()
	if len(reporters) > 0 {
		m := t.getFullMap()
		for _, reporter := range reporters {
			reporter(m)
		}
	}

	if t.parent == nil {
		tManager.removeTrail(t.id)
	} else {
		tManager.setTrail(t.id, t.parent)
	}
}

func (t *opTrail) getFullMap() GenericMap {
	m := make(GenericMap)
	for k, v := range tManager.globals {
		m[k] = v
	}
	// Locals overwrite globals
	for k, v := range t.data {
		m[k] = v
	}
	return m
}