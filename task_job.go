package maa

type TaskJob struct {
	job          Job
	setParamFunc func(id int64, param string) bool
}

func NewTaskJob(id int64, statusFunc func(id int64) Status, setParamFunc func(id int64, param string) bool) TaskJob {
	job := NewJob(id, statusFunc)
	return TaskJob{
		job:          job,
		setParamFunc: setParamFunc,
	}
}

func (job TaskJob) Status() Status {
	return job.job.Status()
}

func (job TaskJob) Invalid() bool {
	return job.job.Invalid()
}

func (job TaskJob) Pending() bool {
	return job.job.Pending()
}

func (job TaskJob) Running() bool {
	return job.job.Running()
}

func (job TaskJob) Success() bool {
	return job.job.Success()
}

func (job TaskJob) Failure() bool {
	return job.job.Failure()
}

func (job TaskJob) Done() bool {
	return job.job.Done()
}

func (job TaskJob) Wait() bool {
	return job.job.Wait()
}

func (job TaskJob) SetParam(param string) bool {
	return job.setParamFunc(job.job.id, param)
}

func (job TaskJob) GetDetail() (*TaskDetail, bool) {
	return QueryTaskDetail(job.job.id)
}
