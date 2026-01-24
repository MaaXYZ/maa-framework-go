package maa

// Job represents an asynchronous job with status tracking capabilities.
// It provides methods to check the job status and wait for completion.
type Job struct {
	id          int64
	finalStatus Status
	statusFunc  func(id int64) Status
	waitFunc    func(id int64) Status
}

func newJob(id int64, statusFunc func(id int64) Status, waitFunc func(id int64) Status) *Job {
	return &Job{
		id:         id,
		statusFunc: statusFunc,
		waitFunc:   waitFunc,
	}
}

// Status returns the current status of the job.
func (j *Job) Status() Status {
	if j.finalStatus.Invalid() {
		return j.statusFunc(j.id)
	}
	return j.finalStatus
}

// Invalid reports whether the status is invalid.
func (j *Job) Invalid() bool {
	return j.Status().Invalid()
}

// Pending reports whether the status is pending.
func (j *Job) Pending() bool {
	return j.Status().Pending()
}

// Running reports whether the status is running.
func (j *Job) Running() bool {
	return j.Status().Running()
}

// Success reports whether the status is success.
func (j *Job) Success() bool {
	return j.Status().Success()
}

// Failure reports whether the status is a failure.
func (j *Job) Failure() bool {
	return j.Status().Failure()
}

// Done reports whether the job is done (either success or failure).
func (j *Job) Done() bool {
	return j.Status().Done()
}

// Wait blocks until the job completes and returns the job instance.
func (j *Job) Wait() *Job {
	if j.finalStatus.Invalid() {
		j.finalStatus = j.waitFunc(j.id)
	}
	return j
}

// TaskJob extends Job with task-specific functionality.
// It provides additional methods to retrieve task details.
type TaskJob struct {
	*Job
	getTaskDetailFunc    func(id int64) *TaskDetail
	overridePipelineFunc func(id int64, override any) bool
}

func newTaskJob(
	id int64,
	statusFunc func(id int64) Status,
	waitFunc func(id int64) Status,
	getTaskDetailFunc func(id int64) *TaskDetail,
	overridePipelineFunc func(id int64, override any) bool,
) *TaskJob {
	job := newJob(id, statusFunc, waitFunc)
	return &TaskJob{
		Job:                  job,
		getTaskDetailFunc:    getTaskDetailFunc,
		overridePipelineFunc: overridePipelineFunc,
	}
}

// Wait blocks until the task job completes and returns the TaskJob instance.
func (j *TaskJob) Wait() *TaskJob {
	j.Job.Wait()
	return j
}

// GetDetail retrieves the detailed information of the task.
func (j *TaskJob) GetDetail() *TaskDetail {
	return j.getTaskDetailFunc(j.id)
}

// OverridePipeline overrides the pipeline for a running task.
// The `override` parameter can be a JSON string or any data type that can be marshaled to JSON.
func (j *TaskJob) OverridePipeline(override any) bool {
	if j.overridePipelineFunc == nil {
		return false
	}
	return j.overridePipelineFunc(j.id, override)
}
