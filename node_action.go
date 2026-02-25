package maa

import (
	"encoding/json"
	"errors"
	"slices"
	"time"
)

// NodeAction defines the action configuration for a node.
type NodeAction struct {
	// Type specifies the action type.
	Type NodeActionType `json:"type,omitempty"`
	// Param specifies the action parameters.
	Param NodeActionParam `json:"param,omitempty"`
}

func (na *NodeAction) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  NodeActionType  `json:"type,omitempty"`
		Param json.RawMessage `json:"param,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	na.Type = raw.Type

	if len(raw.Param) == 0 || string(raw.Param) == "null" {
		return nil
	}

	var param NodeActionParam
	switch na.Type {
	case NodeActionTypeDoNothing, "":
		param = &NodeDoNothingParam{}
	case NodeActionTypeClick:
		param = &NodeClickParam{}
	case NodeActionTypeLongPress:
		param = &NodeLongPressParam{}
	case NodeActionTypeSwipe:
		param = &NodeSwipeParam{}
	case NodeActionTypeMultiSwipe:
		param = &NodeMultiSwipeParam{}
	case NodeActionTypeTouchDown:
		param = &NodeTouchDownParam{}
	case NodeActionTypeTouchMove:
		param = &NodeTouchMoveParam{}
	case NodeActionTypeTouchUp:
		param = &NodeTouchUpParam{}
	case NodeActionTypeClickKey:
		param = &NodeClickKeyParam{}
	case NodeActionTypeLongPressKey:
		param = &NodeLongPressKeyParam{}
	case NodeActionTypeKeyDown:
		param = &NodeKeyDownParam{}
	case NodeActionTypeKeyUp:
		param = &NodeKeyUpParam{}
	case NodeActionTypeInputText:
		param = &NodeInputTextParam{}
	case NodeActionTypeStartApp:
		param = &NodeStartAppParam{}
	case NodeActionTypeStopApp:
		param = &NodeStopAppParam{}
	case NodeActionTypeStopTask:
		param = &NodeStopTaskParam{}
	case NodeActionTypeScroll:
		param = &NodeScrollParam{}
	case NodeActionTypeCommand:
		param = &NodeCommandParam{}
	case NodeActionTypeShell:
		param = &NodeShellParam{}
	case NodeActionTypeCustom:
		param = &NodeCustomActionParam{}
	default:
		return errors.New("unsupported action type: " + string(na.Type))
	}

	if err := json.Unmarshal(raw.Param, param); err != nil {
		return err
	}
	na.Param = param
	return nil
}

// NodeActionType defines the available action types.
type NodeActionType string

const (
	NodeActionTypeDoNothing    NodeActionType = "DoNothing"
	NodeActionTypeClick        NodeActionType = "Click"
	NodeActionTypeLongPress    NodeActionType = "LongPress"
	NodeActionTypeSwipe        NodeActionType = "Swipe"
	NodeActionTypeMultiSwipe   NodeActionType = "MultiSwipe"
	NodeActionTypeTouchDown    NodeActionType = "TouchDown"
	NodeActionTypeTouchMove    NodeActionType = "TouchMove"
	NodeActionTypeTouchUp      NodeActionType = "TouchUp"
	NodeActionTypeClickKey     NodeActionType = "ClickKey"
	NodeActionTypeLongPressKey NodeActionType = "LongPressKey"
	NodeActionTypeKeyDown      NodeActionType = "KeyDown"
	NodeActionTypeKeyUp        NodeActionType = "KeyUp"
	NodeActionTypeInputText    NodeActionType = "InputText"
	NodeActionTypeStartApp     NodeActionType = "StartApp"
	NodeActionTypeStopApp      NodeActionType = "StopApp"
	NodeActionTypeStopTask     NodeActionType = "StopTask"
	NodeActionTypeScroll       NodeActionType = "Scroll"
	NodeActionTypeCommand      NodeActionType = "Command"
	NodeActionTypeShell        NodeActionType = "Shell"
	NodeActionTypeCustom       NodeActionType = "Custom"
)

// NodeActionParam is the interface for action parameters.
type NodeActionParam interface {
	isActionParam()
}

// NodeDoNothingParam defines parameters for do-nothing action.
type NodeDoNothingParam struct{}

func (n NodeDoNothingParam) isActionParam() {}

// ActDoNothing creates a DoNothing action that performs no operation.
func ActDoNothing() *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeDoNothing,
		Param: &NodeDoNothingParam{},
	}
}

// NodeClickParam defines parameters for click action.
type NodeClickParam struct {
	// Target specifies the click target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeClickParam) isActionParam() {}

// ActClick creates a Click action. All fields are optional; pass no argument for defaults.
func ActClick(p ...NodeClickParam) *NodeAction {
	var param NodeClickParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeClick, Param: &param}
}

// NodeLongPressParam defines parameters for long press action.
type NodeLongPressParam struct {
	// Target specifies the long press target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Duration specifies the long press duration. Default: 1000ms.
	// JSON: serialized as integer milliseconds.
	Duration time.Duration `json:"-"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeLongPressParam) isActionParam() {}

func (p NodeLongPressParam) MarshalJSON() ([]byte, error) {
	type NoMethod NodeLongPressParam
	return json.Marshal(struct {
		NoMethod
		Duration int64 `json:"duration,omitempty"`
	}{NoMethod: NoMethod(p), Duration: p.Duration.Milliseconds()})
}

func (p *NodeLongPressParam) UnmarshalJSON(data []byte) error {
	type NoMethod NodeLongPressParam
	raw := struct {
		NoMethod
		Duration int64 `json:"duration,omitempty"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = NodeLongPressParam(raw.NoMethod)
	p.Duration = time.Duration(raw.Duration) * time.Millisecond
	return nil
}

// ActLongPress creates a LongPress action. All fields are optional; pass no argument for defaults.
func ActLongPress(p ...NodeLongPressParam) *NodeAction {
	var param NodeLongPressParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeLongPress, Param: &param}
}

// NodeSwipeParam defines parameters for swipe action.
type NodeSwipeParam struct {
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	BeginOffset Rect `json:"begin_offset,omitempty"`
	// End specifies the swipe end position.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration. Default: 200ms.
	// JSON: serialized as array of integer milliseconds.
	Duration []time.Duration `json:"-"`
	// EndHold specifies extra wait time at end position before releasing. Default: 0.
	// JSON: serialized as array of integer milliseconds.
	EndHold []time.Duration `json:"-"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeSwipeParam) isActionParam() {}

func (p NodeSwipeParam) MarshalJSON() ([]byte, error) {
	type NoMethod NodeSwipeParam
	return json.Marshal(struct {
		NoMethod
		Duration []int64 `json:"duration,omitempty"`
		EndHold  []int64 `json:"end_hold,omitempty"`
	}{NoMethod: NoMethod(p), Duration: durationsToMs(p.Duration), EndHold: durationsToMs(p.EndHold)})
}

func (p *NodeSwipeParam) UnmarshalJSON(data []byte) error {
	type NoMethod NodeSwipeParam
	raw := struct {
		NoMethod
		Duration []int64 `json:"duration,omitempty"`
		EndHold  []int64 `json:"end_hold,omitempty"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = NodeSwipeParam(raw.NoMethod)
	p.Duration = msToDurations(raw.Duration)
	p.EndHold = msToDurations(raw.EndHold)
	return nil
}

// ActSwipe creates a Swipe action. All fields are optional; pass no argument for defaults.
func ActSwipe(p ...NodeSwipeParam) *NodeAction {
	var param NodeSwipeParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeSwipe, Param: &param}
}

// NodeMultiSwipeItem defines a single swipe within a multi-swipe action.
type NodeMultiSwipeItem struct {
	// Starting specifies when this swipe starts within the action. Default: 0.
	// JSON: serialized as integer milliseconds.
	Starting time.Duration `json:"-"`
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	BeginOffset Rect `json:"begin_offset,omitempty"`
	// End specifies the swipe end position.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration. Default: 200ms.
	// JSON: serialized as array of integer milliseconds.
	Duration []time.Duration `json:"-"`
	// EndHold specifies extra wait time at end position before releasing. Default: 0.
	// JSON: serialized as array of integer milliseconds.
	EndHold []time.Duration `json:"-"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index. Win32: mouse button. Default uses array index if 0.
	Contact int `json:"contact,omitempty"`
}

func (p NodeMultiSwipeItem) MarshalJSON() ([]byte, error) {
	type NoMethod NodeMultiSwipeItem
	return json.Marshal(struct {
		NoMethod
		Starting int64   `json:"starting,omitempty"`
		Duration []int64 `json:"duration,omitempty"`
		EndHold  []int64 `json:"end_hold,omitempty"`
	}{
		NoMethod: NoMethod(p),
		Starting: p.Starting.Milliseconds(),
		Duration: durationsToMs(p.Duration),
		EndHold:  durationsToMs(p.EndHold),
	})
}

func (p *NodeMultiSwipeItem) UnmarshalJSON(data []byte) error {
	type NoMethod NodeMultiSwipeItem
	raw := struct {
		NoMethod
		Starting int64   `json:"starting,omitempty"`
		Duration []int64 `json:"duration,omitempty"`
		EndHold  []int64 `json:"end_hold,omitempty"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = NodeMultiSwipeItem(raw.NoMethod)
	p.Starting = time.Duration(raw.Starting) * time.Millisecond
	p.Duration = msToDurations(raw.Duration)
	p.EndHold = msToDurations(raw.EndHold)
	return nil
}

// NodeMultiSwipeParam defines parameters for multi-finger swipe action.
type NodeMultiSwipeParam struct {
	// Swipes specifies the list of swipe items. Required.
	Swipes []NodeMultiSwipeItem `json:"swipes,omitempty"`
}

func (n NodeMultiSwipeParam) isActionParam() {}

// ActMultiSwipe creates a MultiSwipe action for multi-finger swipe gestures.
func ActMultiSwipe(swipes ...NodeMultiSwipeItem) *NodeAction {
	param := &NodeMultiSwipeParam{
		Swipes: swipes,
	}
	return &NodeAction{
		Type:  NodeActionTypeMultiSwipe,
		Param: param,
	}
}

// NodeTouchDownParam defines parameters for touch down action.
type NodeTouchDownParam struct {
	// Target specifies the touch target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Pressure specifies the touch pressure, range depends on controller implementation. Default: 0.
	Pressure int `json:"pressure,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchDownParam) isActionParam() {}

// ActTouchDown creates a TouchDown action. All fields are optional; pass no argument for defaults.
func ActTouchDown(p ...NodeTouchDownParam) *NodeAction {
	var param NodeTouchDownParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeTouchDown, Param: &param}
}

// NodeTouchMoveParam defines parameters for touch move action.
type NodeTouchMoveParam struct {
	// Target specifies the touch target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Pressure specifies the touch pressure, range depends on controller implementation. Default: 0.
	Pressure int `json:"pressure,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchMoveParam) isActionParam() {}

// ActTouchMove creates a TouchMove action. All fields are optional; pass no argument for defaults.
func ActTouchMove(p ...NodeTouchMoveParam) *NodeAction {
	var param NodeTouchMoveParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeTouchMove, Param: &param}
}

// NodeTouchUpParam defines parameters for touch up action.
type NodeTouchUpParam struct {
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchUpParam) isActionParam() {}

// ActTouchUp creates a TouchUp action. All fields are optional; pass no argument for defaults.
func ActTouchUp(p ...NodeTouchUpParam) *NodeAction {
	var param NodeTouchUpParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeTouchUp, Param: &param}
}

// NodeClickKeyParam defines parameters for key click action.
type NodeClickKeyParam struct {
	// Key specifies the virtual key codes to click. Required.
	Key []int `json:"key,omitempty"`
}

func (n NodeClickKeyParam) isActionParam() {}

// ActClickKey creates a ClickKey action with the given virtual key codes.
func ActClickKey(keys []int) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeClickKey,
		Param: &NodeClickKeyParam{Key: slices.Clone(keys)},
	}
}

// NodeLongPressKeyParam defines parameters for long press key action.
type NodeLongPressKeyParam struct {
	// Key specifies the virtual key code to press. Required.
	Key []int `json:"key,omitempty"`
	// Duration specifies the long press duration. Default: 1000ms.
	// JSON: serialized as integer milliseconds.
	Duration time.Duration `json:"-"`
}

func (n NodeLongPressKeyParam) isActionParam() {}

func (p NodeLongPressKeyParam) MarshalJSON() ([]byte, error) {
	type NoMethod NodeLongPressKeyParam
	return json.Marshal(struct {
		NoMethod
		Duration int64 `json:"duration,omitempty"`
	}{NoMethod: NoMethod(p), Duration: p.Duration.Milliseconds()})
}

func (p *NodeLongPressKeyParam) UnmarshalJSON(data []byte) error {
	type NoMethod NodeLongPressKeyParam
	raw := struct {
		NoMethod
		Duration int64 `json:"duration,omitempty"`
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = NodeLongPressKeyParam(raw.NoMethod)
	p.Duration = time.Duration(raw.Duration) * time.Millisecond
	return nil
}

// ActLongPressKey creates a LongPressKey action with the given parameters.
func ActLongPressKey(p NodeLongPressKeyParam) *NodeAction {
	return &NodeAction{Type: NodeActionTypeLongPressKey, Param: &p}
}

// NodeKeyDownParam defines parameters for key down action.
type NodeKeyDownParam struct {
	// Key specifies the virtual key code to press down. Required.
	Key int `json:"key,omitempty"`
}

func (n NodeKeyDownParam) isActionParam() {}

// ActKeyDown creates a KeyDown action that presses the key without releasing.
func ActKeyDown(key int) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeKeyDown,
		Param: &NodeKeyDownParam{Key: key},
	}
}

// NodeKeyUpParam defines parameters for key up action.
type NodeKeyUpParam struct {
	// Key specifies the virtual key code to release. Required.
	Key int `json:"key,omitempty"`
}

func (n NodeKeyUpParam) isActionParam() {}

// ActKeyUp creates a KeyUp action that releases a previously pressed key.
func ActKeyUp(key int) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeKeyUp,
		Param: &NodeKeyUpParam{Key: key},
	}
}

// NodeInputTextParam defines parameters for text input action.
type NodeInputTextParam struct {
	// InputText specifies the text to input. Some controllers only support ASCII. Required.
	InputText string `json:"input_text,omitempty"`
}

func (n NodeInputTextParam) isActionParam() {}

// ActInputText creates an InputText action with the given text.
func ActInputText(input string) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeInputText,
		Param: &NodeInputTextParam{InputText: input},
	}
}

// NodeStartAppParam defines parameters for start app action.
type NodeStartAppParam struct {
	// Package specifies the package name or activity to start. Required.
	Package string `json:"package,omitempty"`
}

func (n NodeStartAppParam) isActionParam() {}

// ActStartApp creates a StartApp action with the given package name or activity.
func ActStartApp(pkg string) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeStartApp,
		Param: &NodeStartAppParam{Package: pkg},
	}
}

// NodeStopAppParam defines parameters for stop app action.
type NodeStopAppParam struct {
	// Package specifies the package name to stop. Required.
	Package string `json:"package,omitempty"`
}

func (n NodeStopAppParam) isActionParam() {}

// ActStopApp creates a StopApp action with the given package name.
func ActStopApp(pkg string) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeStopApp,
		Param: &NodeStopAppParam{Package: pkg},
	}
}

// NodeStopTaskParam defines parameters for stop task action.
// This action stops the current task chain.
type NodeStopTaskParam struct{}

func (n NodeStopTaskParam) isActionParam() {}

// ActStopTask creates a StopTask action that stops the current task chain.
func ActStopTask() *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeStopTask,
		Param: &NodeStopTaskParam{},
	}
}

// NodeScrollParam defines parameters for scroll action.
type NodeScrollParam struct {
	// Target specifies the scroll target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Dx specifies the horizontal scroll amount.
	Dx int `json:"dx,omitempty"`
	// Dy specifies the vertical scroll amount.
	Dy int `json:"dy,omitempty"`
}

func (n NodeScrollParam) isActionParam() {}

// ActScroll creates a Scroll action. All fields are optional; pass no argument for defaults.
func ActScroll(p ...NodeScrollParam) *NodeAction {
	var param NodeScrollParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeAction{Type: NodeActionTypeScroll, Param: &param}
}

// NodeCommandParam defines parameters for command execution action.
type NodeCommandParam struct {
	// Exec specifies the program path to execute. Required.
	Exec string `json:"exec,omitempty"`
	// Args specifies the command arguments. Supports runtime placeholders:
	// {ENTRY}: task entry name, {NODE}: current node name,
	// {IMAGE}: screenshot file path, {BOX}: recognition target [x,y,w,h],
	// {RESOURCE_DIR}: last loaded resource directory, {LIBRARY_DIR}: MaaFW library directory.
	Args []string `json:"args,omitempty"`
	// Detach enables detached mode to run without waiting for completion. Default: false.
	Detach bool `json:"detach,omitempty"`
}

func (n NodeCommandParam) isActionParam() {}

// ActCommand creates a Command action with the given parameters.
func ActCommand(p NodeCommandParam) *NodeAction {
	return &NodeAction{Type: NodeActionTypeCommand, Param: &p}
}

// NodeShellParam defines parameters for shell command execution action.
type NodeShellParam struct {
	Cmd string `json:"cmd,omitempty"`
}

func (n NodeShellParam) isActionParam() {}

// ActShell creates a Shell action with the given command.
// This is only valid for ADB controllers. If the controller is not an ADB controller, the action will fail.
// The output of the command can be obtained in the action detail by MaaTaskerGetActionDetail.
func ActShell(cmd string) *NodeAction {
	return &NodeAction{Type: NodeActionTypeShell, Param: &NodeShellParam{Cmd: cmd}}
}

// NodeCustomActionParam defines parameters for custom action handlers.
type NodeCustomActionParam struct {
	// Target specifies the action target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// CustomAction specifies the action name registered via MaaResourceRegisterCustomAction. Required.
	CustomAction string `json:"custom_action,omitempty"`
	// CustomActionParam specifies custom parameters passed to the action callback.
	CustomActionParam any `json:"custom_action_param,omitempty"`
}

func (n NodeCustomActionParam) isActionParam() {}

// ActCustom creates a Custom action with the given parameters.
func ActCustom(p NodeCustomActionParam) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeCustom,
		Param: &p,
	}
}
