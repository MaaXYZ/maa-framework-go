package maa

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

type testNotificationHandlerOnRawNotification struct {
	*NotificationHandler
}

func (t *testNotificationHandlerOnRawNotification) OnRawNotification(msg, detailsJson string) {
	fmt.Printf("TestNotificationHandler_OnRawNotification, msg: %s, detailsJson: %s\n", msg, detailsJson)
	t.NotificationHandler.OnRawNotification(msg, detailsJson)
}

func TestNotificationHandler_OnRawNotification(t *testing.T) {
	ctrl := createDbgController(t, &testNotificationHandlerOnRawNotification{})
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
