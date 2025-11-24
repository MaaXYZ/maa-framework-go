type {{.InterfaceType}} interface {
    OnResourceLoading({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail ResourceLoadingDetail)
    OnControllerAction({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail ControllerActionDetail)
    OnTaskerTask({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail TaskerTaskDetail)
    OnNodePipelineNode({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail NodePipelineNodeDetail)
    OnNodeRecognitionNode({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail NodeRecognitionNodeDetail)
    OnNodeActionNode({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail NodeActionNodeDetail)
    OnTaskNextList({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail NodeNextListDetail)
    OnTaskRecognition({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail NodeRecognitionDetail)
    OnTaskAction({{.ReceiverName}} {{.ReceiverType}}, event EventStatus, detail NodeActionDetail)
    OnUnknownEvent({{.ReceiverName}} {{.ReceiverType}}, msg, detailsJSON string)
}
