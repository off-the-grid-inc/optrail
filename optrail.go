package optrail

import (
	"fmt"
	"sync"
	"sync/atomic"

	atomic2 "github.com/uber-go/atomic"
)

var (
	tManager = newTrailManager()
	rManager = newReporterManager()
)

// Begin creates a new OpTrail
func Begin(name string) OpTrail {
	return tManager.beginOpTrail().Here("name", name)
}

func PutGlobal(key string, value interface{}) {
	tManager.putGlobal(key, value)
}

func CurrentTrail() *opTrail {
	return tManager.getTrail(curGoroutineID())
}

// OpTrail is an operation that can cross API boundaries, representation formats, storage
// and tranmission, and remembers its history
type OpTrail interface {
	// Succeed marks this OpTrail as succeeded
	Succeed()

	// MaybeFail fails the OpTrail immediately if the given err is not nil
	MaybeFail(err error) error

	// Here marks this point in time and adds a specific key value to provide
	// information about it
	Here(key string, value interface{}) OpTrail

	// Vanish makes the trail disappear, no more actions can be done with it
	Vanish()

	// Fork creates a child OpTrail
	Fork() OpTrail

	// Transmute lets the current OpTrail to continue temporarily as a different goroutine
	Transmute(fn func())
}

type opTrail struct {
	sync.RWMutex
	id        goId
	parent    *opTrail
	spawnFrom *opTrail
	data      TimestampedMap
	finalized *atomic2.Bool
}

func makeOpTrail(id goId, parent *opTrail, spawnFrom *opTrail) *opTrail {
	return &opTrail{
		id:        id,
		parent:    parent,
		spawnFrom: spawnFrom,
		data:      make(TimestampedMap),
		finalized: atomic2.NewBool(false),
	}
}

func (t *opTrail) Succeed() {
	if t.finalized.Load() {
		return
	}

	t.Lock()
	t.data["succeeded"] = newTimestamped(true)
	t.Unlock()
	t.finalize(true)
	return
}

func (t *opTrail) MaybeFail(err error) error {
	if t.finalized.Load() {
		return err
	}

	t.Lock()
	vals := newTimestampedMulti(false, err)
	t.data["succeeded"] = vals[0]
	t.data["error"] = vals[1]
	t.Unlock()
	t.finalize(true)
	return err
}

func (t *opTrail) Here(key string, value interface{}) OpTrail {
	id := atomic.LoadUint64(&t.id)
	if curGoroutineID() != id {
		panic("calling from the wrong goroutine")
	}
	t.Lock()
	t.data[key] = newTimestamped(value)
	t.Unlock()
	return t
}

func (t *opTrail) Transmute(fn func()) {
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

func (t *opTrail) Fork() OpTrail {
	id := atomic.LoadUint64(&t.id)
	next := makeOpTrail(id, t, nil)
	tManager.Lock()
	tManager.trails[id] = next
	tManager.Unlock()
	return next
}

func (t *opTrail) Vanish() {
	t.Lock()
	t.data = make(TimestampedMap)
	t.Unlock()
	t.finalize(false)
}

func (t *opTrail) finalize(doReport bool) {
	if doReport {
		reporters := rManager.getReporters()
		if len(reporters) > 0 {
			m := t.getFullMap()
			for _, reporter := range reporters {
				reporter(m)
			}
		}
	}

	id := atomic.LoadUint64(&t.id)
	if t.parent == nil {
		tManager.removeTrail(id)
	} else {
		tManager.setTrail(id, t.parent)
	}
	t.finalized.Store(true)
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

func printMap(m GenericMap) {
	for k, v := range m {
		switch v := v.(type) {
		case *timestamped:
			fmt.Printf("T[%v] %v: %v\n", v.timestamp, k, v.data)
		default:
			fmt.Printf("[GLOBAL] %v: %v\n", k, v)
		}
	}
}
