package maa

// Status represents the lifecycle state of a task or item.
type Status int32

const (
	// StatusInvalid is the unknown or uninitialized state, matching
	// MaaStatus_Invalid in the native API. It is the zero value.
	StatusInvalid Status = 0
	// StatusPending means the item is queued but not yet started, matching
	// MaaStatus_Pending.
	StatusPending Status = 1000
	// StatusRunning means work is in progress, matching MaaStatus_Running.
	StatusRunning Status = 2000
	// StatusSuccess means the item completed successfully, matching
	// MaaStatus_Succeeded.
	StatusSuccess Status = 3000
	// StatusFailure means the item completed with failure, matching
	// MaaStatus_Failed.
	StatusFailure Status = 4000
)

// Invalid reports whether the status is StatusInvalid.
func (s Status) Invalid() bool {
	return s == StatusInvalid
}

// Pending reports whether the status is StatusPending.
func (s Status) Pending() bool {
	return s == StatusPending
}

// Running reports whether the status is StatusRunning.
func (s Status) Running() bool {
	return s == StatusRunning
}

// Success reports whether the status is StatusSuccess.
func (s Status) Success() bool {
	return s == StatusSuccess
}

// Failure reports whether the status is StatusFailure.
func (s Status) Failure() bool {
	return s == StatusFailure
}

// Done reports whether the status is terminal (success or failure).
func (s Status) Done() bool {
	return s.Success() || s.Failure()
}

// String returns the human-readable representation of the Status. Unknown
// values render as "invalid", while Invalid only reports true for the exact
// StatusInvalid value.
func (s Status) String() string {
	switch s {
	case StatusInvalid:
		return "invalid"
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusFailure:
		return "failure"
	default:
		return "invalid"
	}
}
