package maa

type Status int32

const (
	Invalid Status = 0
	Pending Status = 1000
	Running Status = 2000
	Success Status = 3000
	Failed  Status = 4000
)
