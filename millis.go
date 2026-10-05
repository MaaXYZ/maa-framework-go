package maa

import (
	"errors"
	"time"
)

// durationFromMs converts integer milliseconds, rejecting values outside the
// range of time.Duration.
func durationFromMs(ms int64) (time.Duration, error) {
	const maxMilliseconds = int64((1<<63 - 1) / int64(time.Millisecond))
	if ms > maxMilliseconds || ms < -maxMilliseconds {
		return 0, errors.New("duration exceeds time.Duration range")
	}
	return time.Duration(ms) * time.Millisecond, nil
}

func durationsToMs(ds []time.Duration) []int64 {
	if ds == nil {
		return nil
	}
	ms := make([]int64, len(ds))
	for i, d := range ds {
		ms[i] = d.Milliseconds()
	}
	return ms
}

func msToDurations(ms []int64) []time.Duration {
	if ms == nil {
		return nil
	}
	ds := make([]time.Duration, len(ms))
	for i, m := range ms {
		ds[i] = time.Duration(m) * time.Millisecond
	}
	return ds
}
