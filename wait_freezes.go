package maa

import (
	"bytes"
	"errors"
	"time"
)

// WaitFreezesParam defines parameters for waiting until screen stabilizes.
// The screen is considered stable when there are no significant changes for a continuous period.
// JSON input may be a bare number, shorthand for the time in milliseconds.
type WaitFreezesParam struct {
	// Time specifies the duration that the screen must remain stable.
	// Zero is omitted, inheriting the existing value; the framework's built-in default is zero (no wait).
	// JSON: serialized as integer milliseconds.
	Time time.Duration `json:"-"`
	// Target specifies the region to monitor for changes.
	// The built-in default is the current recognition box.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// Threshold specifies the template matching threshold for detecting changes. Default: 0.95.
	Threshold float64 `json:"threshold,omitempty"`
	// Method specifies the template matching algorithm (cv::TemplateMatchModes). Default: 5.
	Method int `json:"method,omitempty"`
	// RateLimit specifies the minimum interval between checks. Default: 1000ms.
	// JSON: serialized as integer milliseconds.
	RateLimit time.Duration `json:"-"`
	// Timeout specifies the maximum wait time. Default: 20000ms.
	// JSON: serialized as integer milliseconds.
	Timeout time.Duration `json:"-"`
}

// MarshalJSON encodes time, rate_limit, and timeout as integer milliseconds.
// Zero durations are omitted so pipeline defaults and parent-node
// inheritance apply.
func (w WaitFreezesParam) MarshalJSON() ([]byte, error) {
	type NoMethod WaitFreezesParam
	return marshalJSON(struct {
		NoMethod
		Time      int64 `json:"time,omitempty"`
		RateLimit int64 `json:"rate_limit,omitempty"`
		Timeout   int64 `json:"timeout,omitempty"`
	}{
		NoMethod:  NoMethod(w),
		Time:      w.Time.Milliseconds(),
		RateLimit: w.RateLimit.Milliseconds(),
		Timeout:   w.Timeout.Milliseconds(),
	})
}

// UnmarshalJSON decodes integer milliseconds for time, rate_limit, and
// timeout into durations, accepting a bare number as the time in
// milliseconds, matching the native parser. An explicit JSON null is
// rejected, matching the native parser. Invalid input leaves the
// receiver unchanged.
// A bare millisecond value outside the range of time.Duration is rejected.
func (w *WaitFreezesParam) UnmarshalJSON(data []byte) error {
	if shorthand, ok, err := waitFreezesShorthand(data); err != nil {
		return err
	} else if ok {
		*w = WaitFreezesParam{Time: shorthand}
		return nil
	}
	type NoMethod WaitFreezesParam
	raw := struct {
		NoMethod
		Time      int64 `json:"time,omitempty"`
		RateLimit int64 `json:"rate_limit,omitempty"`
		Timeout   int64 `json:"timeout,omitempty"`
	}{}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	*w = WaitFreezesParam(raw.NoMethod)
	w.Time = time.Duration(raw.Time) * time.Millisecond
	w.RateLimit = time.Duration(raw.RateLimit) * time.Millisecond
	w.Timeout = time.Duration(raw.Timeout) * time.Millisecond
	return nil
}

// waitFreezesShorthand decodes the bare-number wait-freezes shorthand to the
// time in milliseconds; ok is false when the input is an object. An explicit
// null is an error, matching the native parser.
func waitFreezesShorthand(data []byte) (time.Duration, bool, error) {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		return 0, false, errors.New("wait freezes must not be null")
	}
	if len(trimmed) == 0 || trimmed[0] == '{' {
		return 0, false, nil
	}
	var ms int64
	if err := unmarshalJSON(data, &ms); err != nil {
		return 0, false, err
	}
	duration, err := durationFromMs(ms)
	return duration, true, err
}
