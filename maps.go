package optrail

import "time"

type GenericMap map[string]interface{}

type TimestampedMap map[string]*timestamped

type timestamped struct {
	timestamp int64
	data      interface{}
}

func newTimestamped(d interface{}) *timestamped {
	return &timestamped{
		timestamp: time.Now().UnixNano(),
		data:      d,
	}
}

func newTimestampedMulti(values ...interface{}) []*timestamped {
	multi := make([]*timestamped, len(values))
	ts := time.Now().UnixNano()
	for i, d := range values {
		multi[i] = &timestamped{
			timestamp: ts,
			data:      d,
		}
	}
	return multi
}
