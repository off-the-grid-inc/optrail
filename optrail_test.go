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

func TestSimple(t *testing.T) {
	require := require.New(t)

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

	require.Error(op.FailIf(errors.New("fake error")), "error didn't pass through")
}
