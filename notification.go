package maa

import (
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	notificationCallbackID          uint64
	notificationCallbackAgents      = make(map[uint64]Notification)
	notificationCallbackAgentsMutex sync.RWMutex
)

func registerNotificationCallback(notify Notification) uint64 {
	id := atomic.AddUint64(&notificationCallbackID, 1)

	notificationCallbackAgentsMutex.Lock()
	notificationCallbackAgents[id] = notify
	notificationCallbackAgentsMutex.Unlock()

	return id
}

func unregisterNotificationCallback(id uint64) {
	notificationCallbackAgentsMutex.Lock()
	delete(notificationCallbackAgents, id)
	notificationCallbackAgentsMutex.Unlock()
}

type NotificationType int

// NotificationType
const (
	NotificationTypeUnknown NotificationType = iota
	NotificationTypeStarting
	NotificationTypeSucceeded
	NotificationTypeFailed
)

type ResourceLoadingDetail struct {
	ResID uint64 `json:"res_id"`
	Hash  string `json:"hash"`
	Path  string `json:"path"`
}

type ControllerActionDetail struct {
	CtrlID uint64 `json:"ctrl_id"`
	UUID   string `json:"uuid"`
	Action string `json:"action"`
}

type TaskerTaskDetail struct {
	TaskID uint64 `json:"task_id"`
	Entry  string `json:"entry"`
	UUID   string `json:"uuid"`
	Hash   string `json:"hash"`
}

type NodeNextListDetail struct {
	TaskID   uint64   `json:"task_id"`
	Name     string   `json:"name"`
	NextList []string `json:"next_list"`
}

type NodeRecognitionDetail struct {
	TaskID uint64 `json:"task_id"`
	RecID  uint64 `json:"reco_id"`
	Name   string `json:"name"`
}

type NodeActionDetail struct {
	TaskID uint64 `json:"task_id"`
	NodeID uint64 `json:"node_id"`
	Name   string `json:"name"`
}

type Notification interface {
	OnResourceLoading(notifyType NotificationType, detail ResourceLoadingDetail)
	OnControllerAction(notifyType NotificationType, detail ControllerActionDetail)
	OnTaskerTask(notifyType NotificationType, detail TaskerTaskDetail)
	OnTaskNextList(notifyType NotificationType, detail NodeNextListDetail)
	OnTaskRecognition(notifyType NotificationType, detail NodeRecognitionDetail)
	OnTaskAction(notifyType NotificationType, detail NodeActionDetail)
	OnUnknownNotification(msg, detailsJSON string)
}

type notificationHandler struct {
	notify Notification
}

func (n *notificationHandler) OnRawNotification(msg, detailsJSON string) {
	if n.notify == nil {
		return
	}

	notifyType := n.notificationType(msg)
	switch {
	case strings.HasPrefix(msg, "Resource.Loading"):
		var detail ResourceLoadingDetail
		_ = formJSON([]byte(detailsJSON), &detail)
		n.notify.OnResourceLoading(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Controller.Action"):
		var detail ControllerActionDetail
		_ = formJSON([]byte(detailsJSON), &detail)
		n.notify.OnControllerAction(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Tasker.Task"):
		var detail TaskerTaskDetail
		_ = formJSON([]byte(detailsJSON), &detail)
		n.notify.OnTaskerTask(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Node.NextList"):
		var detail NodeNextListDetail
		_ = formJSON([]byte(detailsJSON), &detail)
		n.notify.OnTaskNextList(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Node.Recognition"):
		var detail NodeRecognitionDetail
		_ = formJSON([]byte(detailsJSON), &detail)
		n.notify.OnTaskRecognition(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Node.Action"):
		var detail NodeActionDetail
		_ = formJSON([]byte(detailsJSON), &detail)
		n.notify.OnTaskAction(notifyType, detail)
		return
	default:
		n.notify.OnUnknownNotification(msg, detailsJSON)
	}
}

func (n *notificationHandler) notificationType(msg string) NotificationType {
	switch {
	case strings.HasSuffix(msg, ".Starting"):
		return NotificationTypeStarting
	case strings.HasSuffix(msg, ".Succeeded"):
		return NotificationTypeSucceeded
	case strings.HasSuffix(msg, ".Failed"):
		return NotificationTypeFailed
	default:
		return NotificationTypeUnknown
	}
}

// Deprecated: use _MaaEventCallbackAgent instead
func _MaaNotificationCallbackAgent(message, detailsJson *byte, notifyTransArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(notifyTransArg)

	notificationCallbackAgentsMutex.RLock()
	notify, exists := notificationCallbackAgents[id]
	notificationCallbackAgentsMutex.RUnlock()

	if !exists || notify == nil {
		return 0
	}

	handler := &notificationHandler{notify: notify}
	handler.OnRawNotification(bytePtrToString(message), bytePtrToString(detailsJson))
	return 0
}

func _MaaEventCallbackAgent(handle uintptr, message, detailsJson *byte, transArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(transArg)

	notificationCallbackAgentsMutex.RLock()
	notify, exists := notificationCallbackAgents[id]
	notificationCallbackAgentsMutex.RUnlock()

	if !exists || notify == nil {
		return 0
	}

	handler := &notificationHandler{notify: notify}
	handler.OnRawNotification(bytePtrToString(message), bytePtrToString(detailsJson))
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
