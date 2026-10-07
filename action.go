package maa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"
)

// Action defines a node's action using the pipeline v2 type/param object format.
// Known types decode to their typed parameters. Unrecognized type names retain
// their parameter JSON as *RawActionParam; this does not establish native support.
// Unknown fields outside param, and unmodeled fields of known parameters, are not retained.
// Known parameter fields accept only their canonical JSON forms, the shapes
// this package emits: list-typed fields (key, swipe durations and end holds,
// swipe end and end offsets, command args) reject the single-value shorthand
// the native pipeline parser also tolerates, so write the list form.
type Action struct {
	// Type specifies the action type.
	Type ActionType `json:"type,omitempty"`
	// Param specifies the action parameters.
	// A nil Param omits param when encoding. For unknown types, an explicit JSON null is retained.
	Param ActionParam `json:"param,omitempty"`
}

// UnmarshalJSON decodes a pipeline v2 action. Errors in known parameter types
// are returned without falling back to raw JSON. On error the action is unchanged.
// When "param" is absent, parameters are decoded from the whole action object,
// matching the native parser, so flat fields such as
// {"type":"Click","contact":1} are preserved instead of silently lost.
func (na *Action) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  ActionType      `json:"type,omitempty"`
		Param json.RawMessage `json:"param,omitempty"`
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}

	paramData := raw.Param
	if len(paramData) == 0 {
		paramData = data
	}
	param, err := decodeActionParam(raw.Type, paramData)
	if err != nil {
		return err
	}
	*na = Action{Type: raw.Type, Param: param}
	return nil
}

func decodeActionParam(actionType ActionType, data []byte) (ActionParam, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var param ActionParam
	switch actionType {
	case ActionTypeDoNothing, "":
		param = &DoNothingParam{}
	case ActionTypeClick:
		param = &ClickParam{}
	case ActionTypeLongPress:
		param = &LongPressParam{}
	case ActionTypeSwipe:
		param = &SwipeParam{}
	case ActionTypeMultiSwipe:
		param = &MultiSwipeParam{}
	case ActionTypeTouchDown:
		param = &TouchDownParam{}
	case ActionTypeTouchMove:
		param = &TouchMoveParam{}
	case ActionTypeTouchUp:
		param = &TouchUpParam{}
	case ActionTypeClickKey:
		param = &ClickKeyParam{}
	case ActionTypeLongPressKey:
		param = &LongPressKeyParam{}
	case ActionTypeKeyDown:
		param = &KeyDownParam{}
	case ActionTypeKeyUp:
		param = &KeyUpParam{}
	case ActionTypeInputText:
		param = &InputTextParam{}
	case ActionTypeStartApp:
		param = &StartAppParam{}
	case ActionTypeStopApp:
		param = &StopAppParam{}
	case ActionTypeStopTask:
		param = &StopTaskParam{}
	case ActionTypeScroll:
		param = &ScrollParam{}
	case ActionTypeCommand:
		param = &CommandParam{}
	case ActionTypeShell:
		param = &ShellParam{}
	case ActionTypeScreencap:
		param = &ScreencapParam{}
	case ActionTypeCustom:
		param = &CustomActionParam{}
	default:
		param = new(RawActionParam)
	}

	if _, raw := param.(*RawActionParam); !raw && bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil, nil
	}
	if err := unmarshalJSON(data, param); err != nil {
		return nil, err
	}
	return param, nil
}

// ActionType names a pipeline v2 action. Constants identify the types modeled by
// this package; other names may be used with RawActionParam if the native library supports them.
type ActionType string

// Pipeline v2 action types.
const (
	ActionTypeDoNothing    ActionType = "DoNothing"
	ActionTypeClick        ActionType = "Click"
	ActionTypeLongPress    ActionType = "LongPress"
	ActionTypeSwipe        ActionType = "Swipe"
	ActionTypeMultiSwipe   ActionType = "MultiSwipe"
	ActionTypeTouchDown    ActionType = "TouchDown"
	ActionTypeTouchMove    ActionType = "TouchMove"
	ActionTypeTouchUp      ActionType = "TouchUp"
	ActionTypeClickKey     ActionType = "ClickKey"
	ActionTypeLongPressKey ActionType = "LongPressKey"
	ActionTypeKeyDown      ActionType = "KeyDown"
	ActionTypeKeyUp        ActionType = "KeyUp"
	ActionTypeInputText    ActionType = "InputText"
	ActionTypeStartApp     ActionType = "StartApp"
	ActionTypeStopApp      ActionType = "StopApp"
	ActionTypeStopTask     ActionType = "StopTask"
	ActionTypeScroll       ActionType = "Scroll"
	ActionTypeCommand      ActionType = "Command"
	ActionTypeShell        ActionType = "Shell"
	ActionTypeScreencap    ActionType = "Screencap"
	ActionTypeCustom       ActionType = "Custom"
)

// ActionParam is the interface for typed action parameters and RawActionParam.
type ActionParam interface {
	isActionParam()
}

// DoNothingParam defines parameters for do-nothing action.
type DoNothingParam struct{}

func (n DoNothingParam) isActionParam() {}

// ActDoNothing creates a DoNothing action that performs no operation.
func ActDoNothing() *Action {
	return &Action{
		Type:  ActionTypeDoNothing,
		Param: &DoNothingParam{},
	}
}

// ClickParam defines parameters for click action.
type ClickParam struct {
	// Target specifies the click target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
	// Pressure specifies touch pressure; its range depends on the controller.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// Click actions (default_pipeline.json), or the framework default of 1,
	// applies. A pointer to 0 sends zero explicitly.
	Pressure *int `json:"pressure,omitempty"`
}

func (n ClickParam) isActionParam() {}

// ActClick creates a Click action. Pass a zero value for defaults.
func ActClick(p ClickParam) *Action {
	param := p
	return &Action{Type: ActionTypeClick, Param: &param}
}

// LongPressParam defines parameters for long press action.
type LongPressParam struct {
	// Target specifies the long press target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// Duration specifies the long press duration, serialized as integer milliseconds.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// LongPress actions (default_pipeline.json), or the framework default of
	// 1000ms, applies. A pointer to 0 sends zero explicitly.
	Duration *time.Duration `json:"-"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
	// Pressure specifies touch pressure; its range depends on the controller.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// LongPress actions (default_pipeline.json), or the framework default of 1,
	// applies. A pointer to 0 sends zero explicitly.
	Pressure *int `json:"pressure,omitempty"`
}

func (n LongPressParam) isActionParam() {}

// MarshalJSON encodes the long press param, with the duration as integer
// milliseconds. A nil duration omits the field; a pointer to zero is emitted.
func (p LongPressParam) MarshalJSON() ([]byte, error) {
	type NoMethod LongPressParam
	var duration *int64
	if p.Duration != nil {
		ms := p.Duration.Milliseconds()
		duration = &ms
	}
	return marshalJSON(struct {
		NoMethod
		Duration *int64 `json:"duration,omitempty"`
	}{NoMethod: NoMethod(p), Duration: duration})
}

// UnmarshalJSON decodes the long press param, reading the duration as integer
// milliseconds. An absent duration stays nil; zero decodes to a pointer to zero.
// It rejects durations outside the range of time.Duration.
func (p *LongPressParam) UnmarshalJSON(data []byte) error {
	type NoMethod LongPressParam
	raw := struct {
		NoMethod
		Duration *int64 `json:"duration,omitempty"`
	}{}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	decoded := LongPressParam(raw.NoMethod)
	if raw.Duration != nil {
		duration, err := durationFromMs(*raw.Duration)
		if err != nil {
			return err
		}
		decoded.Duration = &duration
	}
	*p = decoded
	return nil
}

// ActLongPress creates a LongPress action. Pass a zero value for defaults.
func ActLongPress(p LongPressParam) *Action {
	param := p
	return &Action{Type: ActionTypeLongPress, Param: &param}
}

// SwipeParam defines parameters for swipe action.
type SwipeParam struct {
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	BeginOffset Rect `json:"begin_offset,omitzero"`
	// End specifies the swipe end positions. JSON input may be a single
	// target in place of the list (a bare true or node name, or a flat
	// [x, y] point or [x, y, w, h] rectangle array); a single target
	// normalizes to a one-element list. A zero Target element fails JSON
	// encoding.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position. JSON
	// input may be a single flat [x, y, w, h] rectangle in place of the
	// list; a single offset normalizes to a one-element list.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration. Default: 200ms.
	// JSON input may be a single number or an array; a single value
	// normalizes to a one-element array.
	// JSON: serialized as array of integer milliseconds.
	Duration []time.Duration `json:"-"`
	// EndHold specifies extra wait time at end position before releasing. Default: 0.
	// JSON input may be a single number or an array; a single value
	// normalizes to a one-element array.
	// JSON: serialized as array of integer milliseconds.
	EndHold []time.Duration `json:"-"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
	// Pressure specifies touch pressure; its range depends on the controller.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// Swipe actions (default_pipeline.json), or the framework default of 1,
	// applies. A pointer to 0 sends zero explicitly.
	Pressure *int `json:"pressure,omitempty"`
}

func (n SwipeParam) isActionParam() {}

// MarshalJSON encodes the swipe param, with the durations and end holds as integer milliseconds.
func (p SwipeParam) MarshalJSON() ([]byte, error) {
	type NoMethod SwipeParam
	return marshalJSON(struct {
		NoMethod
		Duration []int64 `json:"duration,omitempty"`
		EndHold  []int64 `json:"end_hold,omitempty"`
	}{NoMethod: NoMethod(p), Duration: durationsToMs(p.Duration), EndHold: durationsToMs(p.EndHold)})
}

// UnmarshalJSON decodes the swipe param, reading the durations and end holds
// as integer milliseconds, normalizing single end targets and offsets to
// one-element lists and single timing values to one-element arrays.
func (p *SwipeParam) UnmarshalJSON(data []byte) error {
	raw := struct {
		Begin       Target               `json:"begin,omitzero"`
		BeginOffset Rect                 `json:"begin_offset,omitzero"`
		End         orSingleList[Target] `json:"end,omitzero"`
		EndOffset   orSingleList[Rect]   `json:"end_offset,omitempty"`
		Duration    orScalarList[int64]  `json:"duration,omitempty"`
		EndHold     orScalarList[int64]  `json:"end_hold,omitempty"`
		OnlyHover   bool                 `json:"only_hover,omitempty"`
		Contact     int                  `json:"contact,omitempty"`
		Pressure    *int                 `json:"pressure,omitempty"`
	}{}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	*p = SwipeParam{
		Begin:       raw.Begin,
		BeginOffset: raw.BeginOffset,
		End:         raw.End,
		EndOffset:   raw.EndOffset,
		Duration:    msToDurations(raw.Duration),
		EndHold:     msToDurations(raw.EndHold),
		OnlyHover:   raw.OnlyHover,
		Contact:     raw.Contact,
		Pressure:    raw.Pressure,
	}
	return nil
}

// ActSwipe creates a Swipe action. Pass a zero value for defaults.
func ActSwipe(p SwipeParam) *Action {
	param := p
	param.End = slices.Clone(p.End)
	param.EndOffset = slices.Clone(p.EndOffset)
	param.Duration = slices.Clone(p.Duration)
	param.EndHold = slices.Clone(p.EndHold)
	return &Action{Type: ActionTypeSwipe, Param: &param}
}

// MultiSwipeItem defines a single swipe within a multi-swipe action.
type MultiSwipeItem struct {
	// Starting specifies when this swipe starts within the action. Default: 0.
	// JSON: serialized as integer milliseconds.
	Starting time.Duration `json:"-"`
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	BeginOffset Rect `json:"begin_offset,omitzero"`
	// End specifies the swipe end positions. JSON input may be a single
	// target in place of the list (a bare true or node name, or a flat
	// [x, y] point or [x, y, w, h] rectangle array); a single target
	// normalizes to a one-element list. A zero Target element fails JSON
	// encoding.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position. JSON
	// input may be a single flat [x, y, w, h] rectangle in place of the
	// list; a single offset normalizes to a one-element list.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration. Default: 200ms.
	// JSON input may be a single number or an array; a single value
	// normalizes to a one-element array.
	// JSON: serialized as array of integer milliseconds.
	Duration []time.Duration `json:"-"`
	// EndHold specifies extra wait time at end position before releasing. Default: 0.
	// JSON input may be a single number or an array; a single value
	// normalizes to a one-element array.
	// JSON: serialized as array of integer milliseconds.
	EndHold []time.Duration `json:"-"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index. Win32: mouse button. Default uses array index if 0.
	Contact int `json:"contact,omitempty"`
	// Pressure specifies touch pressure; nil uses the default Swipe pressure, initially 1.
	// A pointer to 0 sends zero explicitly.
	Pressure *int `json:"pressure,omitempty"`
}

// MarshalJSON encodes the multi-swipe item, with the starting time, durations,
// and end holds as integer milliseconds.
func (p MultiSwipeItem) MarshalJSON() ([]byte, error) {
	type NoMethod MultiSwipeItem
	return marshalJSON(struct {
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

// UnmarshalJSON decodes the multi-swipe item, reading the starting time,
// durations, and end holds as integer milliseconds, normalizing single end
// targets and offsets to one-element lists and single timing values to
// one-element arrays.
func (p *MultiSwipeItem) UnmarshalJSON(data []byte) error {
	raw := struct {
		Starting    int64                `json:"starting,omitempty"`
		Begin       Target               `json:"begin,omitzero"`
		BeginOffset Rect                 `json:"begin_offset,omitzero"`
		End         orSingleList[Target] `json:"end,omitzero"`
		EndOffset   orSingleList[Rect]   `json:"end_offset,omitempty"`
		Duration    orScalarList[int64]  `json:"duration,omitempty"`
		EndHold     orScalarList[int64]  `json:"end_hold,omitempty"`
		OnlyHover   bool                 `json:"only_hover,omitempty"`
		Contact     int                  `json:"contact,omitempty"`
		Pressure    *int                 `json:"pressure,omitempty"`
	}{}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	*p = MultiSwipeItem{
		Starting:    time.Duration(raw.Starting) * time.Millisecond,
		Begin:       raw.Begin,
		BeginOffset: raw.BeginOffset,
		End:         raw.End,
		EndOffset:   raw.EndOffset,
		Duration:    msToDurations(raw.Duration),
		EndHold:     msToDurations(raw.EndHold),
		OnlyHover:   raw.OnlyHover,
		Contact:     raw.Contact,
		Pressure:    raw.Pressure,
	}
	return nil
}

// MultiSwipeParam defines parameters for multi-finger swipe action.
type MultiSwipeParam struct {
	// Swipes specifies the list of swipe items. Required.
	Swipes []MultiSwipeItem `json:"swipes,omitempty"`
}

func (n MultiSwipeParam) isActionParam() {}

// cloneNodeMultiSwipeItems returns a deep copy of swipes so Param does not share slice backing arrays with the caller.
func cloneNodeMultiSwipeItems(swipes []MultiSwipeItem) []MultiSwipeItem {
	out := make([]MultiSwipeItem, len(swipes))
	for i := range swipes {
		out[i] = MultiSwipeItem{
			Starting:    swipes[i].Starting,
			Begin:       swipes[i].Begin,
			BeginOffset: swipes[i].BeginOffset,
			End:         slices.Clone(swipes[i].End),
			EndOffset:   slices.Clone(swipes[i].EndOffset),
			Duration:    slices.Clone(swipes[i].Duration),
			EndHold:     slices.Clone(swipes[i].EndHold),
			OnlyHover:   swipes[i].OnlyHover,
			Contact:     swipes[i].Contact,
			Pressure:    swipes[i].Pressure,
		}
	}
	return out
}

// ActMultiSwipe creates a MultiSwipe action for multi-finger swipe gestures.
func ActMultiSwipe(swipes ...MultiSwipeItem) *Action {
	param := &MultiSwipeParam{
		Swipes: cloneNodeMultiSwipeItems(swipes),
	}
	return &Action{
		Type:  ActionTypeMultiSwipe,
		Param: param,
	}
}

// TouchDownParam defines parameters for touch down action.
type TouchDownParam struct {
	// AutoUp releases still-held contacts when the task stops, finishes, or the controller is destroyed.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// TouchDown actions (default_pipeline.json), or the framework default of
	// false, applies.
	AutoUp *bool `json:"auto_up,omitempty"`
	// Target specifies the touch target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// Pressure specifies the touch pressure, range depends on controller implementation. Default: 0.
	Pressure int `json:"pressure,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n TouchDownParam) isActionParam() {}

// ActTouchDown creates a TouchDown action. Pass a zero value for defaults.
func ActTouchDown(p TouchDownParam) *Action {
	param := p
	return &Action{Type: ActionTypeTouchDown, Param: &param}
}

// TouchMoveParam defines parameters for touch move action.
type TouchMoveParam struct {
	// AutoUp retains the framework's shared touch parameter for JSON round trips.
	// It only affects TouchDown; TouchMove does not use it. Nil leaves it unspecified.
	AutoUp *bool `json:"auto_up,omitempty"`
	// Target specifies the touch target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// Pressure specifies the touch pressure, range depends on controller implementation. Default: 0.
	Pressure int `json:"pressure,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n TouchMoveParam) isActionParam() {}

// ActTouchMove creates a TouchMove action. Pass a zero value for defaults.
func ActTouchMove(p TouchMoveParam) *Action {
	param := p
	return &Action{Type: ActionTypeTouchMove, Param: &param}
}

// TouchUpParam defines parameters for touch up action.
type TouchUpParam struct {
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n TouchUpParam) isActionParam() {}

// ActTouchUp creates a TouchUp action. contact is the touch point identifier (0 for default).
func ActTouchUp(contact int) *Action {
	return &Action{Type: ActionTypeTouchUp, Param: &TouchUpParam{Contact: contact}}
}

// ClickKeyParam defines parameters for key click action.
type ClickKeyParam struct {
	// Key specifies the virtual key codes to click. Required.
	// JSON input may be a single key code or an array; a single value
	// normalizes to a one-element array.
	Key []int `json:"key,omitempty"`
}

// UnmarshalJSON normalizes a single key code to a one-element array.
func (p *ClickKeyParam) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Key orScalarList[int] `json:"key,omitempty"`
	}
	if err := unmarshalJSON(data, &decoded); err != nil {
		return err
	}
	*p = ClickKeyParam{Key: decoded.Key}
	return nil
}

// orScalarList decodes a protocol value-or-array field: a single value
// normalizes to a one-element list, matching the native parser's
// get_and_check_value_or_array.
type orScalarList[T any] []T

func (l *orScalarList[T]) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*l = nil
		return nil
	}
	if trimmed[0] != '[' {
		var single T
		if err := unmarshalJSON(data, &single); err != nil {
			return err
		}
		*l = orScalarList[T]{single}
		return nil
	}
	var list []T
	if err := unmarshalJSON(data, &list); err != nil {
		return err
	}
	*l = list
	return nil
}

// orSingleList decodes a protocol single-or-list field whose single form is
// itself an array (a flat point/rectangle or pair): a flat array normalizes
// to a one-element list, matching the native parser's single-target branch.
// An array whose first element is an array (or an empty array) is a list.
type orSingleList[T any] []T

func (l *orSingleList[T]) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*l = nil
		return nil
	}
	if trimmed[0] == '[' {
		inner := bytes.TrimSpace(trimmed[1:])
		if len(inner) == 0 || inner[0] == '[' || inner[0] == ']' {
			var list []T
			if err := unmarshalJSON(data, &list); err != nil {
				return err
			}
			*l = list
			return nil
		}
	}
	var single T
	if err := unmarshalJSON(data, &single); err != nil {
		return err
	}
	*l = orSingleList[T]{single}
	return nil
}

func (n ClickKeyParam) isActionParam() {}

// ActClickKey creates a ClickKey action with the given virtual key codes.
func ActClickKey(keys []int) *Action {
	return &Action{
		Type:  ActionTypeClickKey,
		Param: &ClickKeyParam{Key: slices.Clone(keys)},
	}
}

// LongPressKeyParam defines parameters for long press key action.
type LongPressKeyParam struct {
	// Key specifies the virtual key codes to press. Required.
	// JSON input may be a single key code or an array; a single value
	// normalizes to a one-element array.
	Key []int `json:"key,omitempty"`
	// Duration specifies the long press duration, serialized as integer milliseconds.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// LongPressKey actions (default_pipeline.json), or the framework default of
	// 1000ms, applies. A pointer to 0 sends zero explicitly.
	Duration *time.Duration `json:"-"`
}

func (n LongPressKeyParam) isActionParam() {}

// MarshalJSON encodes the long press key param, with the duration as integer
// milliseconds. A nil duration omits the field; a pointer to zero is emitted.
func (p LongPressKeyParam) MarshalJSON() ([]byte, error) {
	type NoMethod LongPressKeyParam
	var duration *int64
	if p.Duration != nil {
		ms := p.Duration.Milliseconds()
		duration = &ms
	}
	return marshalJSON(struct {
		NoMethod
		Duration *int64 `json:"duration,omitempty"`
	}{NoMethod: NoMethod(p), Duration: duration})
}

// UnmarshalJSON decodes the long press key param, reading the duration as
// integer milliseconds and normalizing a single key code to a one-element
// array. An absent duration stays nil; zero decodes to a pointer to zero. It
// rejects durations outside the range of time.Duration.
func (p *LongPressKeyParam) UnmarshalJSON(data []byte) error {
	raw := struct {
		Key      orScalarList[int] `json:"key,omitempty"`
		Duration *int64            `json:"duration,omitempty"`
	}{}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	decoded := LongPressKeyParam{Key: raw.Key}
	if raw.Duration != nil {
		duration, err := durationFromMs(*raw.Duration)
		if err != nil {
			return err
		}
		decoded.Duration = &duration
	}
	*p = decoded
	return nil
}

// ActLongPressKey creates a LongPressKey action with the given parameters.
func ActLongPressKey(p LongPressKeyParam) *Action {
	param := p
	param.Key = slices.Clone(p.Key)
	return &Action{Type: ActionTypeLongPressKey, Param: &param}
}

// KeyDownParam defines parameters for key down action.
type KeyDownParam struct {
	// AutoUp releases still-held keys when the task stops, finishes, or the controller is destroyed.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// KeyDown actions (default_pipeline.json), or the framework default of
	// false, applies.
	AutoUp *bool `json:"auto_up,omitempty"`
	// Key specifies the virtual key code to press down. Required.
	// Decoding also accepts the legacy key_code alias; when both fields are
	// present, key wins. Encoding always writes key.
	Key int `json:"key,omitempty"`
}

func (n KeyDownParam) isActionParam() {}

// UnmarshalJSON decodes the key down param, accepting the legacy key_code
// alias for key. When both fields are present, key wins, matching the
// upstream parser. On error the param is unchanged.
func (p *KeyDownParam) UnmarshalJSON(data []byte) error {
	var raw struct {
		AutoUp *bool           `json:"auto_up,omitempty"`
		Key    json.RawMessage `json:"key,omitempty"`
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	decoded := KeyDownParam{AutoUp: raw.AutoUp}
	key, err := firstPresentKeyJSON(raw.Key, keyCodeAliasJSON(data))
	if err != nil {
		return err
	}
	decoded.Key = key
	*p = decoded
	return nil
}

// keyCodeAliasJSON extracts the legacy key_code member from data, if present.
func keyCodeAliasJSON(data []byte) json.RawMessage {
	var alias struct {
		KeyCode json.RawMessage `json:"key_code,omitempty"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&alias); err != nil {
		return nil
	}
	return alias.KeyCode
}

// firstPresentKeyJSON merges the key and key_code wire fields: the first
// present one wins, mirroring the upstream parser's multi-key lookup.
func firstPresentKeyJSON(key, keyCode json.RawMessage) (int, error) {
	for _, field := range []struct {
		name string
		data json.RawMessage
	}{
		{"key", key},
		{"key_code", keyCode},
	} {
		if len(field.data) == 0 {
			continue
		}
		if bytes.Equal(bytes.TrimSpace(field.data), []byte("null")) {
			return 0, fmt.Errorf("%s must be an integer", field.name)
		}
		var value int
		if err := unmarshalJSON(field.data, &value); err != nil {
			return 0, fmt.Errorf("%s must be an integer", field.name)
		}
		return value, nil
	}
	return 0, nil
}

// ActKeyDown creates a KeyDown action that presses the key without releasing.
func ActKeyDown(key int) *Action {
	return &Action{
		Type:  ActionTypeKeyDown,
		Param: &KeyDownParam{Key: key},
	}
}

// KeyUpParam defines parameters for key up action.
type KeyUpParam struct {
	// AutoUp retains the framework's shared key parameter for JSON round trips.
	// It only affects KeyDown; KeyUp does not use it. Nil leaves it unspecified.
	AutoUp *bool `json:"auto_up,omitempty"`
	// Key specifies the virtual key code to release. Required.
	// Decoding also accepts the legacy key_code alias; when both fields are
	// present, key wins. Encoding always writes key.
	Key int `json:"key,omitempty"`
}

func (n KeyUpParam) isActionParam() {}

// UnmarshalJSON decodes the key up param, accepting the legacy key_code
// alias for key. When both fields are present, key wins, matching the
// upstream parser. On error the param is unchanged.
func (p *KeyUpParam) UnmarshalJSON(data []byte) error {
	var raw struct {
		AutoUp *bool           `json:"auto_up,omitempty"`
		Key    json.RawMessage `json:"key,omitempty"`
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	decoded := KeyUpParam{AutoUp: raw.AutoUp}
	key, err := firstPresentKeyJSON(raw.Key, keyCodeAliasJSON(data))
	if err != nil {
		return err
	}
	decoded.Key = key
	*p = decoded
	return nil
}

// ActKeyUp creates a KeyUp action that releases a previously pressed key.
func ActKeyUp(key int) *Action {
	return &Action{
		Type:  ActionTypeKeyUp,
		Param: &KeyUpParam{Key: key},
	}
}

// InputTextParam defines parameters for text input action.
type InputTextParam struct {
	// InputText specifies the text to input. Some controllers only support ASCII. Required.
	InputText string `json:"input_text,omitempty"`
}

func (n InputTextParam) isActionParam() {}

// ActInputText creates an InputText action with the given text.
func ActInputText(input string) *Action {
	return &Action{
		Type:  ActionTypeInputText,
		Param: &InputTextParam{InputText: input},
	}
}

// StartAppParam defines parameters for start app action.
type StartAppParam struct {
	// Package specifies the package name or activity to start. Required.
	Package string `json:"package,omitempty"`
}

func (n StartAppParam) isActionParam() {}

// ActStartApp creates a StartApp action with the given package name or activity.
func ActStartApp(pkg string) *Action {
	return &Action{
		Type:  ActionTypeStartApp,
		Param: &StartAppParam{Package: pkg},
	}
}

// StopAppParam defines parameters for stop app action.
type StopAppParam struct {
	// Package specifies the package name to stop. Required.
	Package string `json:"package,omitempty"`
}

func (n StopAppParam) isActionParam() {}

// ActStopApp creates a StopApp action with the given package name.
func ActStopApp(pkg string) *Action {
	return &Action{
		Type:  ActionTypeStopApp,
		Param: &StopAppParam{Package: pkg},
	}
}

// StopTaskParam defines parameters for stop task action.
// This action stops the current task chain.
type StopTaskParam struct{}

func (n StopTaskParam) isActionParam() {}

// ActStopTask creates a StopTask action that stops the current task chain.
func ActStopTask() *Action {
	return &Action{
		Type:  ActionTypeStopTask,
		Param: &StopTaskParam{},
	}
}

// ScrollParam defines parameters for scroll action.
type ScrollParam struct {
	// Target specifies the scroll target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// Dx specifies the horizontal scroll amount.
	Dx int `json:"dx,omitempty"`
	// Dy specifies the vertical scroll amount.
	Dy int `json:"dy,omitempty"`
}

func (n ScrollParam) isActionParam() {}

// ActScroll creates a Scroll action. Pass a zero value for defaults.
func ActScroll(p ScrollParam) *Action {
	param := p
	return &Action{Type: ActionTypeScroll, Param: &param}
}

// CommandParam defines parameters for command execution action.
type CommandParam struct {
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

func (n CommandParam) isActionParam() {}

// ActCommand creates a Command action with the given parameters.
func ActCommand(p CommandParam) *Action {
	param := p
	param.Args = slices.Clone(p.Args)
	return &Action{Type: ActionTypeCommand, Param: &param}
}

// ShellParam defines parameters for shell command execution action.
type ShellParam struct {
	// Cmd specifies the command to run in the controller's shell.
	Cmd string `json:"cmd,omitempty"`
	// ShellTimeout limits command execution, serialized as integer milliseconds.
	// Nil omits the field. When overriding a node with the same action type,
	// it inherits the existing value; otherwise the default configured for
	// Shell actions (default_pipeline.json), or the framework default of
	// 20 seconds, applies.
	// A pointer to zero sends zero explicitly; -time.Millisecond means wait indefinitely.
	ShellTimeout *time.Duration `json:"-"`
}

func (n ShellParam) isActionParam() {}

// MarshalJSON encodes the shell timeout in milliseconds, preserving an explicit zero.
func (p ShellParam) MarshalJSON() ([]byte, error) {
	type NoMethod ShellParam
	var timeout *int64
	if p.ShellTimeout != nil {
		ms := p.ShellTimeout.Milliseconds()
		timeout = &ms
	}
	return marshalJSON(struct {
		NoMethod
		ShellTimeout *int64 `json:"shell_timeout,omitempty"`
	}{NoMethod: NoMethod(p), ShellTimeout: timeout})
}

// UnmarshalJSON decodes the shell timeout from integer milliseconds.
// It rejects timeout values outside the range of time.Duration.
func (p *ShellParam) UnmarshalJSON(data []byte) error {
	type NoMethod ShellParam
	raw := struct {
		NoMethod
		ShellTimeout *int64 `json:"shell_timeout,omitempty"`
	}{}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	decoded := ShellParam(raw.NoMethod)
	if raw.ShellTimeout != nil {
		const maxMilliseconds = int64((1<<63 - 1) / int64(time.Millisecond))
		if *raw.ShellTimeout > maxMilliseconds || *raw.ShellTimeout < -maxMilliseconds {
			return errors.New("shell_timeout exceeds time.Duration range")
		}
		duration := time.Duration(*raw.ShellTimeout) * time.Millisecond
		decoded.ShellTimeout = &duration
	}
	*p = decoded
	return nil
}

// ActShell creates a Shell action with the given command.
// To specify a timeout, set the returned action's Param to a ShellParam with ShellTimeout.
// Only controllers that support shell commands (in practice, ADB controllers) can run it;
// other controller types fail.
// The output of the command can be obtained in the action detail by MaaTaskerGetActionDetail.
func ActShell(cmd string) *Action {
	return &Action{Type: ActionTypeShell, Param: &ShellParam{Cmd: cmd}}
}

// ScreencapParam defines parameters for screencap action.
type ScreencapParam struct {
	// Filename specifies screencap filename without extension. Empty means auto-generated by MaaFramework.
	Filename string `json:"filename,omitempty"`
	// Format specifies image format. Optional values: "png", "jpg", "jpeg". Default: "png".
	Format string `json:"format,omitempty"`
	// Quality specifies image quality (0-100), only effective for jpg/jpeg. Default: 100.
	Quality int `json:"quality,omitempty"`
}

func (n ScreencapParam) isActionParam() {}

// ActScreencap creates a Screencap action. Pass a zero value for defaults.
func ActScreencap(p ScreencapParam) *Action {
	param := p
	return &Action{Type: ActionTypeScreencap, Param: &param}
}

// CustomActionParam defines parameters for custom action handlers.
type CustomActionParam struct {
	// Target specifies the action target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	// JSON accepts [x, y], which expands to a 1x1 offset at (x, y), or [x, y, w, h];
	// the zero value omits the field, so the pipeline default or parent-node inheritance applies.
	TargetOffset Rect `json:"target_offset,omitzero"`
	// CustomAction specifies the action name registered via MaaResourceRegisterCustomAction. Required.
	CustomAction string `json:"custom_action,omitempty"`
	// CustomActionParam specifies custom parameters passed to the action callback.
	// The value is passed through verbatim by the framework, so decoding keeps
	// numbers as json.Number to preserve integers beyond float64 precision.
	CustomActionParam any `json:"custom_action_param,omitempty"`
}

func (n CustomActionParam) isActionParam() {}

// UnmarshalJSON decodes the custom action param, keeping the numeric fidelity
// of custom_action_param by decoding its numbers as json.Number, exactly as
// the upstream parser passes the sub-JSON through. On error the param is unchanged.
func (p *CustomActionParam) UnmarshalJSON(data []byte) error {
	var raw struct {
		Target            Target          `json:"target,omitzero"`
		TargetOffset      Rect            `json:"target_offset,omitzero"`
		CustomAction      string          `json:"custom_action,omitempty"`
		CustomActionParam json.RawMessage `json:"custom_action_param,omitempty"`
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	decoded := CustomActionParam{
		Target:       raw.Target,
		TargetOffset: raw.TargetOffset,
		CustomAction: raw.CustomAction,
	}
	if len(raw.CustomActionParam) != 0 {
		var param any
		decoder := json.NewDecoder(bytes.NewReader(raw.CustomActionParam))
		decoder.UseNumber()
		if err := decoder.Decode(&param); err != nil {
			return err
		}
		decoded.CustomActionParam = param
	}
	*p = decoded
	return nil
}

// ActCustom creates a Custom action with the given parameters.
func ActCustom(p CustomActionParam) *Action {
	param := p
	return &Action{
		Type:  ActionTypeCustom,
		Param: &param,
	}
}
