package optrail

import (
	"fmt"
	"sync"

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
	finalized    *atomic2.Bool
}

func makeOpTrail(id uint64, parent *opTrail, branchedFrom *opTrail) *opTrail {
	return &opTrail{
		id:           id,
		parent:       parent,
		branchedFrom: branchedFrom,
		data:         make(TimestampedMap),
		finalized:    atomic2.NewBool(false),
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

func (t *opTrail) FailIf(err error) error {
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

	if t.parent == nil {
		tManager.removeTrail(t.id)
	} else {
		tManager.setTrail(t.id, t.parent)
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
