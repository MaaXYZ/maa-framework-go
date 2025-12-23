package maa

type TaskerEventSink interface {
	OnTaskerTask(tasker *Tasker, event EventStatus, detail TaskerTaskDetail)
}

// TaskerEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type TaskerEventSinkAdapter struct {
	onTaskerTask func(EventStatus, TaskerTaskDetail)
}

func (a *TaskerEventSinkAdapter) OnTaskerTask(tasker *Tasker, status EventStatus, detail TaskerTaskDetail) {
	if a == nil || a.onTaskerTask == nil {
		return
	}
	a.onTaskerTask(status, detail)
}

// OnTaskerTask registers a callback sink that only handles Tasker.Task events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnTaskerTask(fn func(EventStatus, TaskerTaskDetail)) int64 {
	sink := &TaskerEventSinkAdapter{onTaskerTask: fn}
	return t.AddSink(sink)
}

type ResourceEventSink interface {
	OnResourceLoading(res *Resource, event EventStatus, detail ResourceLoadingDetail)
}

// ResourceEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type ResourceEventSinkAdapter struct {
	onResourceLoading func(EventStatus, ResourceLoadingDetail)
}

func (a *ResourceEventSinkAdapter) OnResourceLoading(res *Resource, status EventStatus, detail ResourceLoadingDetail) {
	if a == nil || a.onResourceLoading == nil {
		return
	}
	a.onResourceLoading(status, detail)
}

// OnResourceLoading registers a callback sink that only handles Resource.Loading events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (r *Resource) OnResourceLoading(fn func(EventStatus, ResourceLoadingDetail)) int64 {
	sink := &ResourceEventSinkAdapter{onResourceLoading: fn}
	return r.AddSink(sink)
}

type ContextEventSink interface {
	OnNodePipelineNode(ctx *Context, event EventStatus, detail NodePipelineNodeDetail)
	OnNodeRecognitionNode(ctx *Context, event EventStatus, detail NodeRecognitionNodeDetail)
	OnNodeActionNode(ctx *Context, event EventStatus, detail NodeActionNodeDetail)
	OnNodeNextList(ctx *Context, event EventStatus, detail NodeNextListDetail)
	OnNodeRecognition(ctx *Context, event EventStatus, detail NodeRecognitionDetail)
	OnNodeAction(ctx *Context, event EventStatus, detail NodeActionDetail)
}

// ContextEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type ContextEventSinkAdapter struct {
	onNodePipelineNode    func(*Context, EventStatus, NodePipelineNodeDetail)
	onNodeRecognitionNode func(*Context, EventStatus, NodeRecognitionNodeDetail)
	onNodeActionNode      func(*Context, EventStatus, NodeActionNodeDetail)
	onNodeNextList        func(*Context, EventStatus, NodeNextListDetail)
	onNodeRecognition     func(*Context, EventStatus, NodeRecognitionDetail)
	onNodeAction          func(*Context, EventStatus, NodeActionDetail)
}

func (a *ContextEventSinkAdapter) OnNodePipelineNode(ctx *Context, status EventStatus, detail NodePipelineNodeDetail) {
	if a == nil || a.onNodePipelineNode == nil {
		return
	}
	a.onNodePipelineNode(ctx, status, detail)
}

func (a *ContextEventSinkAdapter) OnNodeRecognitionNode(ctx *Context, status EventStatus, detail NodeRecognitionNodeDetail) {
	if a == nil || a.onNodeRecognitionNode == nil {
		return
	}
	a.onNodeRecognitionNode(ctx, status, detail)
}

func (a *ContextEventSinkAdapter) OnNodeActionNode(ctx *Context, status EventStatus, detail NodeActionNodeDetail) {
	if a == nil || a.onNodeActionNode == nil {
		return
	}
	a.onNodeActionNode(ctx, status, detail)
}

func (a *ContextEventSinkAdapter) OnNodeNextList(ctx *Context, status EventStatus, detail NodeNextListDetail) {
	if a == nil || a.onNodeNextList == nil {
		return
	}
	a.onNodeNextList(ctx, status, detail)
}

func (a *ContextEventSinkAdapter) OnNodeRecognition(ctx *Context, status EventStatus, detail NodeRecognitionDetail) {
	if a == nil || a.onNodeRecognition == nil {
		return
	}
	a.onNodeRecognition(ctx, status, detail)
}

func (a *ContextEventSinkAdapter) OnNodeAction(ctx *Context, status EventStatus, detail NodeActionDetail) {
	if a == nil || a.onNodeAction == nil {
		return
	}
	a.onNodeAction(ctx, status, detail)
}

// OnNodePipelineNodeInContext registers a callback sink that only handles Node.PipelineNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodePipelineNodeInContext(fn func(*Context, EventStatus, NodePipelineNodeDetail)) int64 {
	sink := &ContextEventSinkAdapter{onNodePipelineNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeRecognitionNodeInContext registers a callback sink that only handles Node.RecognitionNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeRecognitionNodeInContext(fn func(*Context, EventStatus, NodeRecognitionNodeDetail)) int64 {
	sink := &ContextEventSinkAdapter{onNodeRecognitionNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeActionNodeInContext registers a callback sink that only handles Node.ActionNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeActionNodeInContext(fn func(*Context, EventStatus, NodeActionNodeDetail)) int64 {
	sink := &ContextEventSinkAdapter{onNodeActionNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeNextListInContext registers a callback sink that only handles Node.NextList events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeNextListInContext(fn func(*Context, EventStatus, NodeNextListDetail)) int64 {
	sink := &ContextEventSinkAdapter{onNodeNextList: fn}
	return t.AddContextSink(sink)
}

// OnNodeRecognitionInContext registers a callback sink that only handles Node.Recognition events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeRecognitionInContext(fn func(*Context, EventStatus, NodeRecognitionDetail)) int64 {
	sink := &ContextEventSinkAdapter{onNodeRecognition: fn}
	return t.AddContextSink(sink)
}

// OnNodeActionInContext registers a callback sink that only handles Node.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeActionInContext(fn func(*Context, EventStatus, NodeActionDetail)) int64 {
	sink := &ContextEventSinkAdapter{onNodeAction: fn}
	return t.AddContextSink(sink)
}

type ControllerEventSink interface {
	OnControllerAction(ctrl *Controller, event EventStatus, detail ControllerActionDetail)
}

// ControllerEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type ControllerEventSinkAdapter struct {
	onControllerAction func(EventStatus, ControllerActionDetail)
}

func (a *ControllerEventSinkAdapter) OnControllerAction(ctrl *Controller, status EventStatus, detail ControllerActionDetail) {
	if a == nil || a.onControllerAction == nil {
		return
	}
	a.onControllerAction(status, detail)
}

// OnControllerAction registers a callback sink that only handles Controller.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (c *Controller) OnControllerAction(fn func(EventStatus, ControllerActionDetail)) int64 {
	sink := &ControllerEventSinkAdapter{onControllerAction: fn}
	return c.AddSink(sink)
}
