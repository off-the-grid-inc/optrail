package optrail

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func TestTrailManager(t *testing.T) {
	require := require.New(t)

	op := tManager.beginOpTrail()
	op2 := tManager.getTrail(op.id)
	require.Equal(op, op2, "should be the same op")

	defer tManager.killAll()
}

func TestSimple(t *testing.T) {
	require := require.New(t)
	defer tManager.killAll()

	op := Begin("test-op")
	op.Here("init", "todo bien")
	op.Here("more", "alles gut")

	iMap := (op.(*opTrail).getFullMap())
	require.Equal(iMap["name"].(*timestamped).data, "test-op")
	require.Equal(iMap["init"].(*timestamped).data, "todo bien")
	require.Equal(iMap["more"].(*timestamped).data, "alles gut")
}

func TestSucceed(t *testing.T) {
	require := require.New(t)
	defer tManager.killAll()

	rManager.RegisterReporter(func(m GenericMap) {
		toFind := map[string]bool{"name": true, "step1": true, "step2": true, "step3": true}
		for k := range m {
			delete(toFind, k)
		}
		require.Equal(0, len(toFind), "didn't expect to find anything here")
	})
	defer rManager.ClearReporters()

	op := Begin("test-op")
	op.Here("step1", "step1-val")
	op.Here("step2", true)
	op.Here("step3", 1234567890)
	op.Succeed()
}

func TestFailIf(t *testing.T) {
	require := require.New(t)
	defer tManager.killAll()

	rManager.RegisterReporter(func(m GenericMap) {
		toFind := map[string]bool{"name": true, "step1": true, "step2": true, "step3": true, "error": true}
		for k := range m {
			delete(toFind, k)
		}
		require.Equal(0, len(toFind), "didn't expect to find anything here")
	})
	defer rManager.ClearReporters()

	op := Begin("test-op")
	op.Here("step1", "step1-val")
	op.Here("step2", true)
	op.Here("step3", 1234567890)

	require.Error(op.MaybeFail(errors.New("fake error")), "error didn't pass through")
}

func TestVanish(t *testing.T) {
	defer tManager.killAll()

	rManager.RegisterReporter(func(m GenericMap) {
		if len(m) > 0 {
			t.FailNow()
		}
	})
	defer rManager.ClearReporters()

	op := Begin("test-op")
	op.Here("step1", "step1-val")
	op.Vanish()
	op.Succeed()
}

func TestGo(t *testing.T) {
	require := require.New(t)
	defer tManager.killAll()

	foundParent := 0
	foundForked := 0

	rManager.RegisterReporter(func(m GenericMap) {
		for k, _ := range m {
			switch k {
			case "parent":
				foundParent++
			case "forked":
				foundForked++
			}
		}
	})
	defer rManager.ClearReporters()

	op := Begin("test-op")
	op2 := op.Fork()

	op.Here("parent", "op1")
	op2.Here("forked", "op2")

	op.Succeed()
	op2.Succeed()

	require.Equal(1, foundParent, "expected to find parent once")
	require.Equal(1, foundForked, "expected to find forked once")
}
