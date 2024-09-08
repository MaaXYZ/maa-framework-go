package maa

type Job struct {
	id         int64
	statusFunc func(id int64) Status
	waitFunc   func(id int64) Status
}

func NewJob(id int64, statusFunc func(id int64) Status, waitFunc func(id int64) Status) Job {
	return Job{
		id:         id,
		statusFunc: statusFunc,
		waitFunc:   waitFunc,
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
	return job.waitFunc(job.id).Success()
}

type TaskJob struct {
	Job
	getTaskDetailFunc func(id int64) (TaskDetail, bool)
}

func NewTaskJob(
	id int64,
	statusFunc func(id int64) Status,
	waitFunc func(id int64) Status,
	getTaskDetailFunc func(id int64) (TaskDetail, bool),
) TaskJob {
	job := NewJob(id, statusFunc, waitFunc)
	return TaskJob{
		Job:               job,
		getTaskDetailFunc: getTaskDetailFunc,
	}
}

func (job TaskJob) GetDetail() (TaskDetail, bool) {
	return job.getTaskDetailFunc(job.id)
}
