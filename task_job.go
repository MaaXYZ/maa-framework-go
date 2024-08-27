package maa

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
