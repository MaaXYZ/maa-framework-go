package maa

import (
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

type eventCallback struct {
	id   uint64
	sink any
}

var (
	lastestEventCallbackID uint64
	eventCallbacks         = make(map[uint64]eventCallback)
	eventCallbacksMutex    sync.RWMutex
)

func registerEventCallback(sink any) uint64 {
	id := atomic.AddUint64(&lastestEventCallbackID, 1)

	eventCallbacksMutex.Lock()
	eventCallbacks[id] = eventCallback{
		id:   id,
		sink: sink,
	}
	eventCallbacksMutex.Unlock()

	return id
}

func unregisterEventCallback(id uint64) {
	eventCallbacksMutex.Lock()
	delete(eventCallbacks, id)
	eventCallbacksMutex.Unlock()
}

// EventStatus represents the current state of an event
type EventStatus int

// Event status constants
const (
	EventStatusUnknown EventStatus = iota
	EventStatusStarting
	EventStatusSucceeded
	EventStatusFailed
)

// ResourceLoadingDetail contains information about resource loading events
type ResourceLoadingDetail struct {
	ResID uint64 `json:"res_id"`
	Hash  string `json:"hash"`
	Path  string `json:"path"`
}

// ControllerActionDetail contains information about controller action events
type ControllerActionDetail struct {
	CtrlID uint64         `json:"ctrl_id"`
	UUID   string         `json:"uuid"`
	Action string         `json:"action"`
	Param  map[string]any `json:"param"`
}

// TaskerTaskDetail contains information about tasker task events
type TaskerTaskDetail struct {
	TaskID uint64 `json:"task_id"`
	Entry  string `json:"entry"`
	UUID   string `json:"uuid"`
	Hash   string `json:"hash"`
}

// NodePipelineNodeDetail contains information about pipeline node events
type NodePipelineNodeDetail struct {
	TaskID uint64 `json:"task_id"`
	NodeID uint64 `json:"node_id"`
	Name   string `json:"name"`
	Focus  any    `json:"focus"`
}

// NodeRecognitionNodeDetail contains information about recognition node events
type NodeRecognitionNodeDetail struct {
	TaskID uint64 `json:"task_id"`
	NodeID uint64 `json:"node_id"`
	Name   string `json:"name"`
	Focus  any    `json:"focus"`
}

// NodeActionNodeDetail contains information about action node events
type NodeActionNodeDetail struct {
	TaskID uint64 `json:"task_id"`
	NodeID uint64 `json:"node_id"`
	Name   string `json:"name"`
	Focus  any    `json:"focus"`
}

// NodeNextListDetail contains information about node next list events
type NodeNextListDetail struct {
	TaskID   uint64         `json:"task_id"`
	Name     string         `json:"name"`
	NextList []NodeNextItem `json:"next_list"`
	Focus    any            `json:"focus"`
}

// NodeRecognitionDetail contains information about node recognition events
type NodeRecognitionDetail struct {
	TaskID        uint64 `json:"task_id"`
	RecognitionID uint64 `json:"reco_id"`
	Name          string `json:"name"`
	Focus         any    `json:"focus"`
}

// NodeActionDetail contains information about node action events
type NodeActionDetail struct {
	TaskID   uint64 `json:"task_id"`
	ActionID uint64 `json:"action_id"`
	Name     string `json:"name"`
	Focus    any    `json:"focus"`
}

type eventHandler struct {
	sink any
}

func (n *eventHandler) handleRaw(handle uintptr, msg, detailsJSON string) {
	if n.sink == nil {
		return
	}

	eventType := n.getEventStatus(msg)
	switch {
	case strings.HasPrefix(msg, "Resource.Loading"):
		var detail ResourceLoadingDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {

		case ResourceEventSink:
			s.OnResourceLoading(&Resource{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Controller.Action"):
		var detail ControllerActionDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ControllerEventSink:
			s.OnControllerAction(&Controller{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Tasker.Task"):
		var detail TaskerTaskDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case TaskerEventSink:
			s.OnTaskerTask(&Tasker{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Node.PipelineNode"):
		var detail NodePipelineNodeDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ContextEventSink:
			s.OnNodePipelineNode(&Context{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Node.RecognitionNode"):
		var detail NodeRecognitionNodeDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ContextEventSink:
			s.OnNodeRecognitionNode(&Context{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Node.ActionNode"):
		var detail NodeActionNodeDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ContextEventSink:
			s.OnNodeActionNode(&Context{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Node.NextList"):
		var detail NodeNextListDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ContextEventSink:
			s.OnNodeNextList(&Context{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Node.Recognition"):
		var detail NodeRecognitionDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ContextEventSink:
			s.OnNodeRecognition(&Context{handle: handle}, eventType, detail)
		}
		return

	case strings.HasPrefix(msg, "Node.Action"):
		var detail NodeActionDetail
		_ = json.Unmarshal([]byte(detailsJSON), &detail)
		switch s := n.sink.(type) {
		case ContextEventSink:
			s.OnNodeAction(&Context{handle: handle}, eventType, detail)
		}
		return

	default:
		// do nothing
	}
}

func (n *eventHandler) getEventStatus(msg string) EventStatus {
	switch {
	case strings.HasSuffix(msg, ".Starting"):
		return EventStatusStarting
	case strings.HasSuffix(msg, ".Succeeded"):
		return EventStatusSucceeded
	case strings.HasSuffix(msg, ".Failed"):
		return EventStatusFailed
	default:
		return EventStatusUnknown
	}
}

// handle uintptr:
// - Tasker handle for MaaTasker event
// - Resource handle for MaaResource event
// - Controller handle for MaaController event
// - Context handle for MaaContext event
func _MaaEventCallbackAgent(handle uintptr, message, detailsJson *byte, transArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(transArg)

	eventCallbacksMutex.RLock()
	cb, exists := eventCallbacks[id]
	eventCallbacksMutex.RUnlock()

	if !exists || cb.sink == nil {
		return 0
	}

	handler := &eventHandler{
		sink: cb.sink,
	}
	handler.handleRaw(
		handle,
		bytePtrToString(message),
		bytePtrToString(detailsJson),
	)
	return 0
}

func bytePtrToString(b *byte) string {
	length := 0
	for ptr := b; *ptr != 0; ptr = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + 1)) {
		length++
	}
	byteSlice := unsafe.Slice(b, length)

	return string(byteSlice)
}
