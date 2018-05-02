package optrail

import (
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

	iMap := (op.(*opTrail).getFullMap())
	require.Equal(iMap["name"].(*timestamped).data, "test-op")
	require.Equal(iMap["init"].(*timestamped).data, "todo bien")
}
