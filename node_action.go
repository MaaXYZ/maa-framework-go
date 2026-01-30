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

	// If no param provided or null, just return with type set
	if len(raw.Param) == 0 || string(raw.Param) == "null" {
		return nil
	}

	// Unmarshal param based on type
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

// ClickOption is a functional option for configuring NodeClickParam.
type ClickOption func(*NodeClickParam)

// WithClickTarget sets the click target position.
func WithClickTarget(target Target) ClickOption {
	return func(p *NodeClickParam) {
		p.Target = target
	}
}

// WithClickTargetOffset sets additional offset applied to target.
func WithClickTargetOffset(offset Rect) ClickOption {
	return func(p *NodeClickParam) {
		p.TargetOffset = offset
	}
}

// WithClickContact sets the touch point identifier.
func WithClickContact(contact int) ClickOption {
	return func(p *NodeClickParam) {
		p.Contact = contact
	}
}

// ActClick creates a Click action with the given options.
func ActClick(opts ...ClickOption) *NodeAction {
	param := &NodeClickParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeClick, Param: param}
}

// NodeLongPressParam defines parameters for long press action.
type NodeLongPressParam struct {
	// Target specifies the long press target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Duration specifies the long press duration in milliseconds. Default: 1000.
	Duration int64 `json:"duration,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeLongPressParam) isActionParam() {}

// LongPressOption is a functional option for configuring NodeLongPressParam.
type LongPressOption func(*NodeLongPressParam)

// WithLongPressTarget sets the long press target position.
func WithLongPressTarget(target Target) LongPressOption {
	return func(p *NodeLongPressParam) {
		p.Target = target
	}
}

// WithLongPressTargetOffset sets additional offset applied to target.
func WithLongPressTargetOffset(offset Rect) LongPressOption {
	return func(p *NodeLongPressParam) {
		p.TargetOffset = offset
	}
}

// WithLongPressDuration sets the long press duration.
func WithLongPressDuration(d time.Duration) LongPressOption {
	return func(p *NodeLongPressParam) {
		p.Duration = d.Milliseconds()
	}
}

// WithLongPressContact sets the touch point identifier.
func WithLongPressContact(contact int) LongPressOption {
	return func(p *NodeLongPressParam) { p.Contact = contact }
}

// ActLongPress creates a LongPress action with the given options.
func ActLongPress(opts ...LongPressOption) *NodeAction {
	param := &NodeLongPressParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeLongPress, Param: param}
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
	// Duration specifies the swipe duration in milliseconds. Default: 200.
	Duration []int64 `json:"duration,omitempty"`
	// EndHold specifies extra wait time at end position before releasing in milliseconds. Default: 0.
	EndHold []int64 `json:"end_hold,omitempty"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeSwipeParam) isActionParam() {}

// SwipeOption is a functional option for configuring NodeSwipeParam.
type SwipeOption func(*NodeSwipeParam)

// WithSwipeBegin sets the swipe start position.
func WithSwipeBegin(begin Target) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.Begin = begin
	}
}

// WithSwipeBeginOffset sets additional offset applied to begin position.
func WithSwipeBeginOffset(offset Rect) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.BeginOffset = offset
	}
}

// WithSwipeEnd sets the swipe end position.
func WithSwipeEnd(end []Target) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.End = slices.Clone(end)
	}
}

// WithSwipeEndOffset sets additional offset applied to end position.
func WithSwipeEndOffset(offset []Rect) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.EndOffset = slices.Clone(offset)
	}
}

// WithSwipeDuration sets the swipe duration.
func WithSwipeDuration(d []time.Duration) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.Duration = make([]int64, len(d))
		for index, duration := range d {
			p.Duration[index] = duration.Milliseconds()
		}
	}
}

// WithSwipeEndHold sets extra wait time at end position before releasing.
func WithSwipeEndHold(d []time.Duration) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.EndHold = make([]int64, len(d))
		for index, duration := range d {
			p.EndHold[index] = duration.Milliseconds()
		}
	}
}

// WithSwipeOnlyHover enables hover-only mode without press/release actions.
func WithSwipeOnlyHover(only bool) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.OnlyHover = only
	}
}

// WithSwipeContact sets the touch point identifier.
func WithSwipeContact(contact int) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.Contact = contact
	}
}

// ActSwipe creates a Swipe action with the given options.
func ActSwipe(opts ...SwipeOption) *NodeAction {
	param := &NodeSwipeParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{
		Type:  NodeActionTypeSwipe,
		Param: param,
	}
}

// NodeMultiSwipeItem defines a single swipe within a multi-swipe action.
type NodeMultiSwipeItem struct {
	// Starting specifies when this swipe starts within the action in milliseconds. Default: 0.
	Starting int64 `json:"starting,omitempty"`
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	BeginOffset Rect `json:"begin_offset,omitempty"`
	// End specifies the swipe end position.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration in milliseconds. Default: 200.
	Duration []int64 `json:"duration,omitempty"`
	// EndHold specifies extra wait time at end position before releasing in milliseconds. Default: 0.
	EndHold []int64 `json:"end_hold,omitempty"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index. Win32: mouse button. Default uses array index if 0.
	Contact int `json:"contact,omitempty"`
}

// NodeMultiSwipeParam defines parameters for multi-finger swipe action.
type NodeMultiSwipeParam struct {
	// Swipes specifies the list of swipe items. Required.
	Swipes []NodeMultiSwipeItem `json:"swipes,omitempty"`
}

func (n NodeMultiSwipeParam) isActionParam() {}

// MultiSwipeItemOption is a functional option for configuring NodeMultiSwipeItem.
type MultiSwipeItemOption func(*NodeMultiSwipeItem)

// WithMultiSwipeItemStarting sets when this swipe starts within the action.
func WithMultiSwipeItemStarting(starting time.Duration) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Starting = starting.Milliseconds()
	}
}

// WithMultiSwipeItemBegin sets the swipe start position.
func WithMultiSwipeItemBegin(begin Target) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Begin = begin
	}
}

// WithMultiSwipeItemBeginOffset sets additional offset applied to begin position.
func WithMultiSwipeItemBeginOffset(offset Rect) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.BeginOffset = offset
	}
}

// WithMultiSwipeItemEnd sets the swipe end position.
func WithMultiSwipeItemEnd(end []Target) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.End = slices.Clone(end)
	}
}

// WithMultiSwipeItemEndOffset sets additional offset applied to end position.
func WithMultiSwipeItemEndOffset(offset []Rect) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.EndOffset = slices.Clone(offset)
	}
}

// WithMultiSwipeItemDuration sets the swipe duration.
func WithMultiSwipeItemDuration(d []time.Duration) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Duration = make([]int64, len(d))
		for index, duration := range d {
			i.Duration[index] = duration.Milliseconds()
		}
	}
}

// WithMultiSwipeItemEndHold sets extra wait time at end position before releasing.
func WithMultiSwipeItemEndHold(d []time.Duration) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.EndHold = make([]int64, len(d))
		for index, duration := range d {
			i.EndHold[index] = duration.Milliseconds()
		}
	}
}

// WithMultiSwipeItemOnlyHover enables hover-only mode without press/release actions.
func WithMultiSwipeItemOnlyHover(only bool) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.OnlyHover = only
	}
}

// WithMultiSwipeItemContact sets the touch point identifier.
func WithMultiSwipeItemContact(contact int) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Contact = contact
	}
}

// NewMultiSwipeItem creates a new multi-swipe item with the given options.
func NewMultiSwipeItem(opts ...MultiSwipeItemOption) NodeMultiSwipeItem {
	item := NodeMultiSwipeItem{}
	for _, opt := range opts {
		opt(&item)
	}
	return item
}

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

// TouchDownOption is a functional option for configuring NodeTouchDownParam.
type TouchDownOption func(*NodeTouchDownParam)

// WithTouchDownTarget sets the touch target position.
func WithTouchDownTarget(target Target) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.Target = target
	}
}

// WithTouchDownTargetOffset sets additional offset applied to target.
func WithTouchDownTargetOffset(offset Rect) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.TargetOffset = offset
	}
}

// WithTouchDownPressure sets the touch pressure.
func WithTouchDownPressure(pressure int) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.Pressure = pressure
	}
}

// WithTouchDownContact sets the touch point identifier.
func WithTouchDownContact(contact int) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.Contact = contact
	}
}

// ActTouchDown creates a TouchDown action with the given options.
func ActTouchDown(opts ...TouchDownOption) *NodeAction {
	param := &NodeTouchDownParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeTouchDown, Param: param}
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

// TouchMoveOption is a functional option for configuring NodeTouchMoveParam.
type TouchMoveOption func(*NodeTouchMoveParam)

// WithTouchMoveTarget sets the touch target position.
func WithTouchMoveTarget(target Target) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.Target = target
	}
}

// WithTouchMoveTargetOffset sets additional offset applied to target.
func WithTouchMoveTargetOffset(offset Rect) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.TargetOffset = offset
	}
}

// WithTouchMovePressure sets the touch pressure.
func WithTouchMovePressure(pressure int) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.Pressure = pressure
	}
}

// WithTouchMoveContact sets the touch point identifier.
func WithTouchMoveContact(contact int) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.Contact = contact
	}
}

// ActTouchMove creates a TouchMove action with the given options.
func ActTouchMove(opts ...TouchMoveOption) *NodeAction {
	param := &NodeTouchMoveParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeTouchMove, Param: param}
}

// NodeTouchUpParam defines parameters for touch up action.
type NodeTouchUpParam struct {
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchUpParam) isActionParam() {}

// TouchUpOption is a functional option for configuring NodeTouchUpParam.
type TouchUpOption func(*NodeTouchUpParam)

// WithTouchUpContact sets the touch point identifier.
func WithTouchUpContact(contact int) TouchUpOption {
	return func(p *NodeTouchUpParam) {
		p.Contact = contact
	}
}

// ActTouchUp creates a TouchUp action with the given options.
func ActTouchUp(opts ...TouchUpOption) *NodeAction {
	param := &NodeTouchUpParam{}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeAction{Type: NodeActionTypeTouchUp, Param: param}
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
	// Duration specifies the long press duration in milliseconds. Default: 1000.
	Duration int64 `json:"duration,omitempty"`
}

func (n NodeLongPressKeyParam) isActionParam() {}

// LongPressKeyOption is a functional option for configuring NodeLongPressKeyParam.
type LongPressKeyOption func(*NodeLongPressKeyParam)

// WithLongPressKeyDuration sets the long press duration.
func WithLongPressKeyDuration(d time.Duration) LongPressKeyOption {
	return func(p *NodeLongPressKeyParam) { p.Duration = d.Milliseconds() }
}

// ActLongPressKey creates a LongPressKey action with the given virtual key code.
func ActLongPressKey(key []int, opts ...LongPressKeyOption) *NodeAction {
	param := &NodeLongPressKeyParam{
		Key: slices.Clone(key),
	}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeLongPressKey, Param: param}
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

type NodeScrollParam struct {
	Target       Target `json:"target,omitzero"`
	TargetOffset Rect   `json:"target_offset,omitempty"`
	Dx           int    `json:"dx,omitempty"`
	Dy           int    `json:"dy,omitempty"`
}

func (n NodeScrollParam) isActionParam() {}

type ScrollOption func(*NodeScrollParam)

func WithScrollDx(dx int) ScrollOption {
	return func(p *NodeScrollParam) {
		p.Dx = dx
	}
}

func WithScrollDy(dy int) ScrollOption {
	return func(p *NodeScrollParam) {
		p.Dy = dy
	}
}

func ActScroll(opts ...ScrollOption) *NodeAction {
	param := &NodeScrollParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{
		Type:  NodeActionTypeScroll,
		Param: param,
	}
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

// CommandOption is a functional option for configuring NodeCommandParam.
type CommandOption func(*NodeCommandParam)

// WithCommandArgs sets the command arguments.
func WithCommandArgs(args []string) CommandOption {
	return func(p *NodeCommandParam) {
		p.Args = slices.Clone(args)
	}
}

// WithCommandDetach enables detached mode to run without waiting for completion.
func WithCommandDetach(detach bool) CommandOption {
	return func(p *NodeCommandParam) { p.Detach = detach }
}

// ActCommand creates a Command action with the given executable path.
func ActCommand(exec string, opts ...CommandOption) *NodeAction {
	param := &NodeCommandParam{Exec: exec}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeCommand, Param: param}
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

// CustomActionOption is a functional option for configuring NodeCustomActionParam.
type CustomActionOption func(*NodeCustomActionParam)

// WithCustomActionTarget sets the action target position.
func WithCustomActionTarget(target Target) CustomActionOption {
	return func(param *NodeCustomActionParam) {
		param.Target = target
	}
}

// WithCustomActionTargetOffset sets additional offset applied to target.
func WithCustomActionTargetOffset(offset Rect) CustomActionOption {
	return func(param *NodeCustomActionParam) {
		param.TargetOffset = offset
	}
}

// WithCustomActionParam sets custom parameters passed to the action callback.
func WithCustomActionParam(customParam any) CustomActionOption {
	return func(param *NodeCustomActionParam) {
		param.CustomActionParam = customParam
	}
}

// ActCustom creates a Custom action with the given action name.
func ActCustom(name string, opts ...CustomActionOption) *NodeAction {
	param := &NodeCustomActionParam{
		CustomAction: name,
	}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{
		Type:  NodeActionTypeCustom,
		Param: param,
	}
}
