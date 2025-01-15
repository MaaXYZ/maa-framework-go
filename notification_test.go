package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testNotificationHandlerOnRawNotification struct {
}

func (t *testNotificationHandlerOnRawNotification) OnControllerAction(notifyType NotificationType, detail ControllerActionDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnResourceLoading(notifyType NotificationType, detail ResourceLoadingDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskAction(notifyType NotificationType, detail NodeActionDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskNextList(notifyType NotificationType, detail NodeNextListDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskRecognition(notifyType NotificationType, detail NodeRecognitionDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnTaskerTask(notifyType NotificationType, detail TaskerTaskDetail) {
}

func (t *testNotificationHandlerOnRawNotification) OnUnknownNotification(msg string, detailsJSON string) {
}

func newTestNotificationHandlerOnRawNotification() Notification {
	return &testNotificationHandlerOnRawNotification{}
}

func TestNotificationHandler_OnRawNotification(t *testing.T) {
	ctrl := createDbgController(t, newTestNotificationHandlerOnRawNotification())
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t, newTestNotificationHandlerOnRawNotification())
	defer res.Destroy()

	tasker := createTasker(t, newTestNotificationHandlerOnRawNotification())
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	got := tasker.PostTask("TestNotificationHandler_OnRawNotification", J{
		"TestNotificationHandler_OnRawNotification": J{
			"action": "Click",
			"target": []int{100, 200, 100, 100},
			"focus":  true,
		},
	}).Wait().Success()
	require.True(t, got)
}
