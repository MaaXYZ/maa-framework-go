package maa

type Status int32

const (
	StatusInvalid Status = 0
	StatusPending Status = 1000
	StatusRunning Status = 2000
	StatusSuccess Status = 3000
	StatusFailure Status = 4000
)

func (status Status) Invalid() bool {
	return status == StatusInvalid
}

func (status Status) Pending() bool {
	return status == StatusPending
}

func (status Status) Running() bool {
	return status == StatusRunning
}

func (status Status) Success() bool {
	return status == StatusSuccess
}

func (status Status) Failure() bool {
	return status == StatusFailure
}

func (status Status) Done() bool {
	return status.Success() || status.Failure()
}
