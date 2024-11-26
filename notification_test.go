package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testNotificationHandlerOnRawNotification struct {
	t *testing.T
}

func (t *testNotificationHandlerOnRawNotification) OnControllerAction(notifyType NotificationType, detail ControllerActionDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnResourceLoading(notifyType NotificationType, detail ResourceLoadingDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskAction(notifyType NotificationType, detail TaskActionDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskNextList(notifyType NotificationType, detail TaskNextListDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskRecognition(notifyType NotificationType, detail TaskRecognitionDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskerTask(notifyType NotificationType, detail TaskerTaskDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnUnknownNotification(msg string, detailsJSON string) {
}

func NewTestNotificationHandlerOnRawNotification() Notification {
	return &testNotificationHandlerOnRawNotification{}
}

func TestNotificationHandler_OnRawNotification(t *testing.T) {
	ctrl := createDbgController(t, &testNotificationHandlerOnRawNotification{t})
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, &testNotificationHandlerOnRawNotification{})
	defer res.Destroy()

	tasker := createTasker(t, &testNotificationHandlerOnRawNotification{})
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.PostPipeline("TestNotificationHandler_OnRawNotification", J{
		"TestNotificationHandler_OnRawNotification": J{
			"action": "Click",
			"target": []int{100, 200, 100, 100},
		},
	}).Wait().Success()
	require.True(t, got)
}
