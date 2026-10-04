package maa

import (
	"encoding/json"
	"fmt"
)

// Point represents a 2D point [x, y].
type Point [2]int

// X returns the horizontal coordinate of the point.
func (p Point) X() int { return p[0] }

// Y returns the vertical coordinate of the point.
func (p Point) Y() int { return p[1] }

// UnmarshalJSON decodes a point from a two-element array [x, y]. A JSON
// string containing the same array, e.g. `"[1, 2]"`, is also accepted.
func (p *Point) UnmarshalJSON(data []byte) error {
	var raw any
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case string:
		var xy []int
		if err := unmarshalJSON([]byte(v), &xy); err != nil {
			return err
		}
		if len(xy) != 2 {
			return fmt.Errorf("invalid point length: %d", len(xy))
		}
		*p = Point{xy[0], xy[1]}
		return nil
	case []any:
		if len(v) != 2 {
			return fmt.Errorf("invalid point length: %d", len(v))
		}
		x, ok1 := v[0].(float64)
		y, ok2 := v[1].(float64)
		if !ok1 || !ok2 {
			return fmt.Errorf("invalid point element types: %T,%T", v[0], v[1])
		}
		*p = Point{int(x), int(y)}
		return nil
	default:
		return fmt.Errorf("invalid point json type: %T", raw)
	}
}

// ActionResult wraps parsed action detail.
type ActionResult struct {
	tp  ActionType
	val any
}

// Type returns the action type of the result.
func (r *ActionResult) Type() ActionType {
	return r.tp
}

// Value returns the underlying value of the result.
func (r *ActionResult) Value() any {
	return r.val
}

// AsClick returns the click detail and true if the action type is Click.
func (r *ActionResult) AsClick() (*ClickActionResult, bool) {
	if r.tp != ActionTypeClick {
		return nil, false
	}
	val, ok := r.val.(*ClickActionResult)
	return val, ok
}

// AsLongPress returns the long press detail and true if the action type is LongPress.
func (r *ActionResult) AsLongPress() (*LongPressActionResult, bool) {
	if r.tp != ActionTypeLongPress {
		return nil, false
	}
	val, ok := r.val.(*LongPressActionResult)
	return val, ok
}

// AsSwipe returns the swipe detail and true if the action type is Swipe.
func (r *ActionResult) AsSwipe() (*SwipeActionResult, bool) {
	if r.tp != ActionTypeSwipe {
		return nil, false
	}
	val, ok := r.val.(*SwipeActionResult)
	return val, ok
}

// AsMultiSwipe returns the multi-swipe detail and true if the action type is MultiSwipe.
func (r *ActionResult) AsMultiSwipe() (*MultiSwipeActionResult, bool) {
	if r.tp != ActionTypeMultiSwipe {
		return nil, false
	}
	val, ok := r.val.(*MultiSwipeActionResult)
	return val, ok
}

// AsClickKey returns the key detail and true if the action type is ClickKey,
// KeyDown, or KeyUp.
func (r *ActionResult) AsClickKey() (*ClickKeyActionResult, bool) {
	if r.tp != ActionTypeClickKey && r.tp != ActionTypeKeyDown && r.tp != ActionTypeKeyUp {
		return nil, false
	}
	val, ok := r.val.(*ClickKeyActionResult)
	return val, ok
}

// AsLongPressKey returns the long press key detail and true if the action type is LongPressKey.
func (r *ActionResult) AsLongPressKey() (*LongPressKeyActionResult, bool) {
	if r.tp != ActionTypeLongPressKey {
		return nil, false
	}
	val, ok := r.val.(*LongPressKeyActionResult)
	return val, ok
}

// AsInputText returns the input text detail and true if the action type is InputText.
func (r *ActionResult) AsInputText() (*InputTextActionResult, bool) {
	if r.tp != ActionTypeInputText {
		return nil, false
	}
	val, ok := r.val.(*InputTextActionResult)
	return val, ok
}

// AsApp returns the app detail and true if the action type is StartApp or StopApp.
func (r *ActionResult) AsApp() (*AppActionResult, bool) {
	if r.tp != ActionTypeStartApp && r.tp != ActionTypeStopApp {
		return nil, false
	}
	val, ok := r.val.(*AppActionResult)
	return val, ok
}

// AsScroll returns the scroll detail and true if the action type is Scroll.
func (r *ActionResult) AsScroll() (*ScrollActionResult, bool) {
	if r.tp != ActionTypeScroll {
		return nil, false
	}
	val, ok := r.val.(*ScrollActionResult)
	return val, ok
}

// AsTouch returns the touch detail and true if the action type is TouchDown,
// TouchMove, or TouchUp.
func (r *ActionResult) AsTouch() (*TouchActionResult, bool) {
	if r.tp != ActionTypeTouchDown && r.tp != ActionTypeTouchMove && r.tp != ActionTypeTouchUp {
		return nil, false
	}
	val, ok := r.val.(*TouchActionResult)
	return val, ok
}

// AsShell returns the shell detail and true if the action type is Shell.
func (r *ActionResult) AsShell() (*ShellActionResult, bool) {
	if r.tp != ActionTypeShell {
		return nil, false
	}
	val, ok := r.val.(*ShellActionResult)
	return val, ok
}

// AsScreencap returns the screencap detail and true if the action type is Screencap.
func (r *ActionResult) AsScreencap() (*ScreencapActionResult, bool) {
	if r.tp != ActionTypeScreencap {
		return nil, false
	}
	val, ok := r.val.(*ScreencapActionResult)
	return val, ok
}

// ClickActionResult holds the parsed detail of a Click action.
type ClickActionResult struct {
	Point   Point `json:"point"`
	Contact int   `json:"contact"`
	// Pressure is kept to match MaaFramework raw detail JSON.
	Pressure int `json:"pressure"`
}

// LongPressActionResult holds the parsed detail of a LongPress action.
type LongPressActionResult struct {
	Point    Point `json:"point"`
	Duration int64 `json:"duration"`
	Contact  int   `json:"contact"`
	// Pressure is kept to match MaaFramework raw detail JSON.
	Pressure int `json:"pressure"`
}

// SwipeActionResult holds the parsed detail of a Swipe action.
type SwipeActionResult struct {
	Begin     Point   `json:"begin"`
	End       []Point `json:"end"`
	EndHold   []int   `json:"end_hold"`
	Duration  []int   `json:"duration"`
	OnlyHover bool    `json:"only_hover"`
	Starting  int     `json:"starting"`
	Contact   int     `json:"contact"`
	// Pressure is kept to match MaaFramework raw detail JSON.
	Pressure int `json:"pressure"`

	endRaw json.RawMessage
}

type swipeActionResultWire struct {
	Begin     Point           `json:"begin"`
	End       json.RawMessage `json:"end"`
	EndHold   []int           `json:"end_hold"`
	Duration  []int           `json:"duration"`
	OnlyHover bool            `json:"only_hover"`
	Starting  int             `json:"starting"`
	Contact   int             `json:"contact"`
	Pressure  int             `json:"pressure"`
}

// UnmarshalJSON decodes the swipe detail, retaining the raw end JSON so it can
// be re-encoded verbatim.
func (s *SwipeActionResult) UnmarshalJSON(data []byte) error {
	var wire swipeActionResultWire
	if err := unmarshalJSON(data, &wire); err != nil {
		return err
	}

	s.Begin = wire.Begin
	s.EndHold = wire.EndHold
	s.Duration = wire.Duration
	s.OnlyHover = wire.OnlyHover
	s.Starting = wire.Starting
	s.Contact = wire.Contact
	s.Pressure = wire.Pressure

	s.endRaw = append(s.endRaw[:0], wire.End...)

	points, err := parseSwipeEndPoints(wire.End)
	if err != nil {
		return err
	}
	s.End = points
	return nil
}

// MarshalJSON encodes the swipe detail, preferring the retained raw end JSON
// over re-encoding End.
func (s SwipeActionResult) MarshalJSON() ([]byte, error) {
	end := s.endRaw
	if len(end) == 0 {
		// Default JSON representation for end: list of points.
		var err error
		end, err = marshalJSON(s.End)
		if err != nil {
			return nil, err
		}
	}

	return marshalJSON(&swipeActionResultWire{
		Begin:     s.Begin,
		End:       end,
		EndHold:   s.EndHold,
		Duration:  s.Duration,
		OnlyHover: s.OnlyHover,
		Starting:  s.Starting,
		Contact:   s.Contact,
		Pressure:  s.Pressure,
	})
}

func parseSwipeEndPoints(end json.RawMessage) ([]Point, error) {
	// Accepted forms of the swipe end JSON:
	// - array of points: [[x,y], ...]
	// - single point: [x,y]
	// - JSON string of a single point: "[x, y]"
	// Parse into []Point, but preserve original end JSON for marshaling.
	var raw any
	if err := unmarshalJSON(end, &raw); err != nil {
		return nil, err
	}

	switch v := raw.(type) {
	case string:
		// v should be a JSON array string: "[x,y]" or "[[x,y],...]"
		return parseSwipeEndPoints(json.RawMessage([]byte(v)))
	case []any:
		if len(v) == 0 {
			return []Point{}, nil
		}
		// Try: [x,y]
		if _, ok := v[0].(float64); ok {
			if len(v) != 2 {
				return nil, fmt.Errorf("invalid swipe end point length: %d", len(v))
			}
			x, ok1 := v[0].(float64)
			y, ok2 := v[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("invalid swipe end point element types: %T,%T", v[0], v[1])
			}
			return []Point{{int(x), int(y)}}, nil
		}

		// Try: [[x,y], ...]
		points := make([]Point, 0, len(v))
		for _, item := range v {
			arr, ok := item.([]any)
			if !ok || len(arr) != 2 {
				return nil, fmt.Errorf("invalid swipe end element: %T", item)
			}
			x, ok1 := arr[0].(float64)
			y, ok2 := arr[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("invalid swipe end point element types: %T,%T", arr[0], arr[1])
			}
			points = append(points, Point{int(x), int(y)})
		}
		return points, nil
	default:
		return nil, fmt.Errorf("invalid swipe end json type: %T", raw)
	}
}

// MultiSwipeActionResult holds the parsed detail of a MultiSwipe action.
type MultiSwipeActionResult struct {
	Swipes []SwipeActionResult `json:"swipes"`
}

// ClickKeyActionResult holds the parsed detail of ClickKey, KeyDown, and KeyUp actions.
type ClickKeyActionResult struct {
	Keycode []int `json:"keycode"`
	AutoUp  bool  `json:"auto_up"`
}

// LongPressKeyActionResult holds the parsed detail of a LongPressKey action.
type LongPressKeyActionResult struct {
	Keycode  []int `json:"keycode"`
	Duration int64 `json:"duration"`
}

// InputTextActionResult holds the parsed detail of an InputText action.
type InputTextActionResult struct {
	Text string `json:"text"`
}

// AppActionResult holds the parsed detail of StartApp and StopApp actions.
type AppActionResult struct {
	Package string `json:"package"`
}

// ScrollActionResult holds the parsed detail of a Scroll action.
type ScrollActionResult struct {
	// Point is kept to match MaaFramework raw detail JSON.
	Point Point `json:"point"`
	Dx    int   `json:"dx"`
	Dy    int   `json:"dy"`
}

// TouchActionResult holds the parsed detail of TouchDown, TouchMove, and TouchUp actions.
type TouchActionResult struct {
	Contact  int   `json:"contact"`
	Point    Point `json:"point"`
	Pressure int   `json:"pressure"`
	AutoUp   bool  `json:"auto_up"`
}

// ShellActionResult holds the parsed detail of a Shell action.
type ShellActionResult struct {
	Cmd          string `json:"cmd"`
	ShellTimeout int    `json:"shell_timeout"`
	Success      bool   `json:"success"`
	Output       string `json:"output"`
}

// ScreencapActionResult holds the parsed detail of a Screencap action.
type ScreencapActionResult struct {
	Filepath string `json:"filepath"`
	Format   string `json:"format"`
	Quality  int    `json:"quality"`
	Success  bool   `json:"success"`
}

func parseActionResult(action, detailJson string) (*ActionResult, error) {
	if detailJson == "" || detailJson == "{}" || detailJson == "null" {
		return nil, nil
	}

	actionType := ActionType(action)
	var resultVal any
	switch actionType {
	case ActionTypeClick:
		resultVal = &ClickActionResult{}
	case ActionTypeLongPress:
		resultVal = &LongPressActionResult{}
	case ActionTypeSwipe:
		resultVal = &SwipeActionResult{}
	case ActionTypeMultiSwipe:
		resultVal = &MultiSwipeActionResult{}
	case ActionTypeClickKey, ActionTypeKeyDown, ActionTypeKeyUp:
		resultVal = &ClickKeyActionResult{}
	case ActionTypeLongPressKey:
		resultVal = &LongPressKeyActionResult{}
	case ActionTypeInputText:
		resultVal = &InputTextActionResult{}
	case ActionTypeStartApp, ActionTypeStopApp:
		resultVal = &AppActionResult{}
	case ActionTypeScroll:
		resultVal = &ScrollActionResult{}
	case ActionTypeTouchDown, ActionTypeTouchMove, ActionTypeTouchUp:
		resultVal = &TouchActionResult{}
	case ActionTypeShell:
		resultVal = &ShellActionResult{}
	case ActionTypeScreencap:
		resultVal = &ScreencapActionResult{}
	default:
		return nil, fmt.Errorf("unknown action result type: %s", action)
	}

	if err := unmarshalJSON([]byte(detailJson), resultVal); err != nil {
		return nil, err
	}

	return &ActionResult{
		tp:  actionType,
		val: resultVal,
	}, nil
}
