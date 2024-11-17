package maa

import (
	"strings"
	"sync/atomic"
	"unsafe"
)

var (
	notificationCallbackID     uint64
	notificationCallbackAgents = make(map[uint64]Notification)
)

func registerNotificationCallback(notify Notification) uint64 {
	id := atomic.AddUint64(&notificationCallbackID, 1)
	notificationCallbackAgents[id] = notify
	return id
}

func unregisterNotificationCallback(id uint64) {
	delete(notificationCallbackAgents, id)
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

type TaskNextListDetail struct {
	TaskID   uint64   `json:"task_id"`
	Name     string   `json:"name"`
	NextList []string `json:"next_list"`
}

type TaskRecognitionDetail struct {
	TaskID uint64 `json:"task_id"`
	RecID  uint64 `json:"reco_id"`
	Name   string `json:"name"`
}

type TaskActionDetail struct {
	TaskID uint64 `json:"task_id"`
	NodeID uint64 `json:"node_id"`
	Name   string `json:"name"`
}

type Notification interface {
	OnResourceLoading(notifyType NotificationType, detail ResourceLoadingDetail)
	OnControllerAction(notifyType NotificationType, detail ControllerActionDetail)
	OnTaskerTask(notifyType NotificationType, detail TaskerTaskDetail)
	OnTaskNextList(notifyType NotificationType, detail TaskNextListDetail)
	OnTaskRecognition(notifyType NotificationType, detail TaskRecognitionDetail)
	OnTaskAction(notifyType NotificationType, detail TaskActionDetail)
	OnRawNotification(msg, detailsJSON string)
	OnUnknownNotification(msg, detailsJSON string)
}

type NotificationHandler struct{}

func _NewNotificationHandler() Notification {
	return &NotificationHandler{}
}

func (n *NotificationHandler) OnResourceLoading(_ NotificationType, _ ResourceLoadingDetail) {
	// DO NOTHING
}

func (n *NotificationHandler) OnControllerAction(_ NotificationType, _ ControllerActionDetail) {
	// DO NOTHING
}

func (n *NotificationHandler) OnTaskerTask(_ NotificationType, _ TaskerTaskDetail) {
	// DO NOTHING
}

func (n *NotificationHandler) OnTaskNextList(_ NotificationType, _ TaskNextListDetail) {
	// DO NOTHING
}

func (n *NotificationHandler) OnTaskRecognition(_ NotificationType, _ TaskRecognitionDetail) {
	// DO NOTHING
}

func (n *NotificationHandler) OnTaskAction(_ NotificationType, _ TaskActionDetail) {
	// DO NOTHING
}

func (n *NotificationHandler) OnRawNotification(msg, detailsJSON string) {
	notifyType := n.notificationType(msg)
	switch {
	case strings.HasPrefix(msg, "Resource.Loading"):
		var detail ResourceLoadingDetail
		_ = formJSON([]byte(msg), &detail)
		n.OnResourceLoading(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Controller.Action"):
		var detail ControllerActionDetail
		_ = formJSON([]byte(msg), &detail)
		n.OnControllerAction(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Tasker.Task"):
		var detail TaskerTaskDetail
		_ = formJSON([]byte(msg), &detail)
		n.OnTaskerTask(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Task.NextList"):
		var detail TaskNextListDetail
		_ = formJSON([]byte(msg), &detail)
		n.OnTaskNextList(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Task.Recognition"):
		var detail TaskRecognitionDetail
		_ = formJSON([]byte(msg), &detail)
		n.OnTaskRecognition(notifyType, detail)
		return
	case strings.HasPrefix(msg, "Task.Action"):
		var detail TaskActionDetail
		_ = formJSON([]byte(msg), &detail)
		n.OnTaskAction(notifyType, detail)
		return
	default:
		n.OnUnknownNotification(msg, detailsJSON)
	}
}

func (n *NotificationHandler) OnUnknownNotification(_, _ string) {
	// DO NOTHING
}

func (n *NotificationHandler) notificationType(msg string) NotificationType {
	switch {
	case strings.HasSuffix(msg, ".Starting"):
		return NotificationTypeStarting
	case strings.HasSuffix(msg, "Succeeded"):
		return NotificationTypeSucceeded
	case strings.HasSuffix(msg, "Failed"):
		return NotificationTypeFailed
	default:
		return NotificationTypeUnknown
	}
}

func _MaaNotificationCallbackAgent(message, detailsJson *byte, notifyTransArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(notifyTransArg)
	notify := notificationCallbackAgents[id]
	if notify == nil {
		return 0
	}
	notify.OnRawNotification(bytePtrToString(message), bytePtrToString(detailsJson))
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
