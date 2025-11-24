{{define "CommonEventSinkHelper"}}
// OnResourceLoading registers a callback sink that only handles Resource.Loading events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnResourceLoading(fn func(EventStatus, ResourceLoadingDetail)) int64 {
	sink := &{{.AdapterType}}{onResourceLoading: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnControllerAction registers a callback sink that only handles Controller.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnControllerAction(fn func(EventStatus, ControllerActionDetail)) int64 {
	sink := &{{.AdapterType}}{onControllerAction: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnTaskerTask registers a callback sink that only handles Tasker.Task events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnTaskerTask(fn func(EventStatus, TaskerTaskDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskerTask: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnNodePipelineNode registers a callback sink that only handles Node.PipelineNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnNodePipelineNode(fn func(EventStatus, NodePipelineNodeDetail)) int64 {
	sink := &{{.AdapterType}}{onNodePipelineNode: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnNodeRecognitionNode registers a callback sink that only handles Node.RecognitionNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnNodeRecognitionNode(fn func(EventStatus, NodeRecognitionNodeDetail)) int64 {
	sink := &{{.AdapterType}}{onNodeRecognitionNode: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnNodeActionNode registers a callback sink that only handles Node.ActionNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnNodeActionNode(fn func(EventStatus, NodeActionNodeDetail)) int64 {
	sink := &{{.AdapterType}}{onNodeActionNode: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnTaskNextList registers a callback sink that only handles Node.NextList events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnTaskNextList(fn func(EventStatus, NodeNextListDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskNextList: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnTaskRecognition registers a callback sink that only handles Node.Recognition events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnTaskRecognition(fn func(EventStatus, NodeRecognitionDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskRecognition: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnTaskAction registers a callback sink that only handles Node.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnTaskAction(fn func(EventStatus, NodeActionDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskAction: fn}
	return {{.InstanceName}}.AddSink(sink)
}

// OnUnknownEvent registers a callback sink that only handles unknown events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func ({{.InstanceName}} {{.ReceiverType}}) OnUnknownEvent(fn func(msg, detailsJSON string)) int64 {
	sink := &{{.AdapterType}}{onUnknownEvent: fn}
	return {{.InstanceName}}.AddSink(sink)
}
{{end}}

{{define "ContextEventSinkHelper"}}
// OnResourceLoadingInContext  registers a callback sink that only handles Resource.Loading events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnResourceLoadingInContext (fn func({{.ReceiverType}}, EventStatus, ResourceLoadingDetail)) int64 {
	sink := &{{.AdapterType}}{onResourceLoading: fn}
	return t.AddContextSink(sink)
}

// OnControllerActionInContext registers a callback sink that only handles Controller.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnControllerActionInContext(fn func({{.ReceiverType}}, EventStatus, ControllerActionDetail)) int64 {
	sink := &{{.AdapterType}}{onControllerAction: fn}
	return t.AddContextSink(sink)
}

// OnTaskerTaskInContext  registers a callback sink that only handles Tasker.Task events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnTaskerTaskInContext(fn func({{.ReceiverType}}, EventStatus, TaskerTaskDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskerTask: fn}
	return t.AddContextSink(sink)
}

// OnNodePipelineNodeInContext registers a callback sink that only handles Node.PipelineNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodePipelineNodeInContext(fn func({{.ReceiverType}}, EventStatus, NodePipelineNodeDetail)) int64 {
	sink := &{{.AdapterType}}{onNodePipelineNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeRecognitionNodeInContext registers a callback sink that only handles Node.RecognitionNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeRecognitionNodeInContext(fn func({{.ReceiverType}}, EventStatus, NodeRecognitionNodeDetail)) int64 {
	sink := &{{.AdapterType}}{onNodeRecognitionNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeActionNodeInContext registers a callback sink that only handles Node.ActionNode events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnNodeActionNodeInContext(fn func({{.ReceiverType}}, EventStatus, NodeActionNodeDetail)) int64 {
	sink := &{{.AdapterType}}{onNodeActionNode: fn}
	return t.AddContextSink(sink)
}

// OnTaskNextListInContext registers a callback sink that only handles Node.NextList events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnTaskNextListInContext(fn func({{.ReceiverType}}, EventStatus, NodeNextListDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskNextList: fn}
	return t.AddContextSink(sink)
}

// OnTaskRecognitionInContext registers a callback sink that only handles Node.Recognition events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnTaskRecognitionInContext(fn func({{.ReceiverType}}, EventStatus, NodeRecognitionDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskRecognition: fn}
	return t.AddContextSink(sink)
}

// OnTaskActionInContext registers a callback sink that only handles Node.Action events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnTaskActionInContext(fn func({{.ReceiverType}}, EventStatus, NodeActionDetail)) int64 {
	sink := &{{.AdapterType}}{onTaskAction: fn}
	return t.AddContextSink(sink)
}

// OnUnknownEventInContext registers a callback sink that only handles unknown events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) OnUnknownEventInContext(fn func(ctx {{.ReceiverType}}, msg, detailsJSON string)) int64 {
	sink := &{{.AdapterType}}{onUnknownEvent: fn}
	return t.AddContextSink(sink)
}
{{end}}


{{if .ContextAware}}
{{template "ContextEventSinkHelper" .}}
{{else}}
{{template "CommonEventSinkHelper" .}}
{{end}}