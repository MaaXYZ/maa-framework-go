{{define "CommonEventSinkAdapter"}}
// {{.AdapterType}} is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type {{.AdapterType}} struct {
	onResourceLoading     func(EventStatus, ResourceLoadingDetail)
	onControllerAction    func(EventStatus, ControllerActionDetail)
	onTaskerTask          func(EventStatus, TaskerTaskDetail)
	onNodePipelineNode    func(EventStatus, NodePipelineNodeDetail)
	onNodeRecognitionNode func(EventStatus, NodeRecognitionNodeDetail)
	onNodeActionNode      func(EventStatus, NodeActionNodeDetail)
	onTaskNextList        func(EventStatus, NodeNextListDetail)
	onTaskRecognition     func(EventStatus, NodeRecognitionDetail)
	onTaskAction          func(EventStatus, NodeActionDetail)
	onUnknownEvent        func(string, string)
}

func (a *{{.AdapterType}}) OnResourceLoading({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail ResourceLoadingDetail) {
	if a == nil || a.onResourceLoading == nil {
		return
	}
	a.onResourceLoading(status, detail)
}

func (a *{{.AdapterType}}) OnControllerAction({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail ControllerActionDetail) {
	if a == nil || a.onControllerAction == nil {
		return
	}
	a.onControllerAction(status, detail)
}

func (a *{{.AdapterType}}) OnTaskerTask({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail TaskerTaskDetail) {
	if a == nil || a.onTaskerTask == nil {
		return
	}
	a.onTaskerTask(status, detail)
}

func (a *{{.AdapterType}}) OnNodePipelineNode({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodePipelineNodeDetail) {
	if a == nil || a.onNodePipelineNode == nil {
		return
	}
	a.onNodePipelineNode(status, detail)
}

func (a *{{.AdapterType}}) OnNodeRecognitionNode({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeRecognitionNodeDetail) {
	if a == nil || a.onNodeRecognitionNode == nil {
		return
	}
	a.onNodeRecognitionNode(status, detail)
}

func (a *{{.AdapterType}}) OnNodeActionNode({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeActionNodeDetail) {
	if a == nil || a.onNodeActionNode == nil {
		return
	}
	a.onNodeActionNode(status, detail)
}

func (a *{{.AdapterType}}) OnTaskNextList({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeNextListDetail) {
	if a == nil || a.onTaskNextList == nil {
		return
	}
	a.onTaskNextList(status, detail)
}

func (a *{{.AdapterType}}) OnTaskRecognition({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeRecognitionDetail) {
	if a == nil || a.onTaskRecognition == nil {
		return
	}
	a.onTaskRecognition(status, detail)
}

func (a *{{.AdapterType}}) OnTaskAction({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeActionDetail) {
	if a == nil || a.onTaskAction == nil {
		return
	}
	a.onTaskAction(status, detail)
}

func (a *{{.AdapterType}}) OnUnknownEvent({{.ReceiverName}} {{.ReceiverType}}, msg, detailsJSON string) {
	if a == nil || a.onUnknownEvent == nil {
		return
	}
	a.onUnknownEvent(msg, detailsJSON)
}
{{end}}

{{define "ContextEventSinkAdapter"}}
// {{.AdapterType}} is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type {{.AdapterType}} struct {
	onResourceLoading     func({{.ReceiverType}}, EventStatus, ResourceLoadingDetail)
	onControllerAction    func({{.ReceiverType}}, EventStatus, ControllerActionDetail)
	onTaskerTask          func({{.ReceiverType}}, EventStatus, TaskerTaskDetail)
	onNodePipelineNode    func({{.ReceiverType}}, EventStatus, NodePipelineNodeDetail)
	onNodeRecognitionNode func({{.ReceiverType}}, EventStatus, NodeRecognitionNodeDetail)
	onNodeActionNode      func({{.ReceiverType}}, EventStatus, NodeActionNodeDetail)
	onTaskNextList        func({{.ReceiverType}}, EventStatus, NodeNextListDetail)
	onTaskRecognition     func({{.ReceiverType}}, EventStatus, NodeRecognitionDetail)
	onTaskAction          func({{.ReceiverType}}, EventStatus, NodeActionDetail)
	onUnknownEvent        func({{.ReceiverType}}, string, string)
}

func (a *{{.AdapterType}}) OnResourceLoading({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail ResourceLoadingDetail) {
	if a == nil || a.onResourceLoading == nil {
		return
	}
	a.onResourceLoading({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnControllerAction({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail ControllerActionDetail) {
	if a == nil || a.onControllerAction == nil {
		return
	}
	a.onControllerAction({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnTaskerTask({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail TaskerTaskDetail) {
	if a == nil || a.onTaskerTask == nil {
		return
	}
	a.onTaskerTask({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnNodePipelineNode({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodePipelineNodeDetail) {
	if a == nil || a.onNodePipelineNode == nil {
		return
	}
	a.onNodePipelineNode({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnNodeRecognitionNode({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeRecognitionNodeDetail) {
	if a == nil || a.onNodeRecognitionNode == nil {
		return
	}
	a.onNodeRecognitionNode({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnNodeActionNode({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeActionNodeDetail) {
	if a == nil || a.onNodeActionNode == nil {
		return
	}
	a.onNodeActionNode({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnTaskNextList({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeNextListDetail) {
	if a == nil || a.onTaskNextList == nil {
		return
	}
	a.onTaskNextList({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnTaskRecognition({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeRecognitionDetail) {
	if a == nil || a.onTaskRecognition == nil {
		return
	}
	a.onTaskRecognition({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnTaskAction({{.ReceiverName}} {{.ReceiverType}}, status EventStatus, detail NodeActionDetail) {
	if a == nil || a.onTaskAction == nil {
		return
	}
	a.onTaskAction({{.ReceiverName}}, status, detail)
}

func (a *{{.AdapterType}}) OnUnknownEvent({{.ReceiverName}} {{.ReceiverType}}, msg, detailsJSON string) {
	if a == nil || a.onUnknownEvent == nil {
		return
	}
	a.onUnknownEvent({{.ReceiverName}}, msg, detailsJSON)
}
{{end}}

{{if .ContextAware}}
{{template "ContextEventSinkAdapter" .}}
{{else}}
{{template "CommonEventSinkAdapter" .}}
{{end}}