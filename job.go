package maa

import "time"

type Job struct {
	id         int64
	statusFunc func(id int64) Status
}

func NewJob(id int64, statusFunc func(id int64) Status) Job {
	return Job{
		id:         id,
		statusFunc: statusFunc,
	}
}

func (job Job) Status() Status {
	return job.statusFunc(job.id)
}

func (job Job) Invalid() bool {
	return job.Status().Invalid()
}

func (job Job) Pending() bool {
	return job.Status().Pending()
}

func (job Job) Running() bool {
	return job.Status().Running()
}

func (job Job) Success() bool {
	return job.Status().Success()
}

func (job Job) Failure() bool {
	return job.Status().Failure()
}

func (job Job) Done() bool {
	return job.Status().Done()
}

func (job Job) Wait() bool {
	for !job.Done() {
		time.Sleep(time.Millisecond * 10)
	}
	return job.Success()
}

type TaskJob struct {
	Job
	setParamFunc func(id int64, param string) bool
}

func NewTaskJob(id int64, statusFunc func(id int64) Status, setParamFunc func(id int64, param string) bool) TaskJob {
	job := NewJob(id, statusFunc)
	return TaskJob{
		Job:          job,
		setParamFunc: setParamFunc,
	}
}

func (job TaskJob) SetParam(param string) bool {
	return job.setParamFunc(job.id, param)
}

func (job TaskJob) GetDetail() (TaskDetail, bool) {
	return QueryTaskDetail(job.id)
}
