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

func (j Job) Status() Status {
	return j.statusFunc(j.id)
}

func (j Job) Invalid() bool {
	return j.Status().Invalid()
}

func (j Job) Pending() bool {
	return j.Status().Pending()
}

func (j Job) Running() bool {
	return j.Status().Running()
}

func (j Job) Success() bool {
	return j.Status().Success()
}

func (j Job) Failure() bool {
	return j.Status().Failure()
}

func (j Job) Done() bool {
	return j.Status().Done()
}

func (j Job) Wait() Job {
	j.waitFunc(j.id)
	return j
}

type TaskJob struct {
	Job
	getTaskDetailFunc func(id int64) *TaskDetail
}

func NewTaskJob(
	id int64,
	statusFunc func(id int64) Status,
	waitFunc func(id int64) Status,
	getTaskDetailFunc func(id int64) *TaskDetail,
) TaskJob {
	job := NewJob(id, statusFunc, waitFunc)
	return TaskJob{
		Job:               job,
		getTaskDetailFunc: getTaskDetailFunc,
	}
}

func (j TaskJob) Wait() TaskJob {
	job := j.Job.Wait()
	return TaskJob{
		Job:               job,
		getTaskDetailFunc: j.getTaskDetailFunc,
	}
}

func (j TaskJob) GetDetail() *TaskDetail {
	return j.getTaskDetailFunc(j.id)
}
