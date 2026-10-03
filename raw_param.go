package maa

import "encoding/json"

// RawActionParam holds an uninterpreted JSON action parameter value.
// Action.UnmarshalJSON uses *RawActionParam for unrecognized action types.
// It preserves parameter content, including large JSON integers, without converting
// it to Go values. JSON encoding may normalize whitespace and escape sequences.
//
// It can also be supplied wherever ActionParam is accepted, for example:
//
//	action := &Action{
//		Type:  ActionType("FutureAction"),
//		Param: RawActionParam(`{"new_option":true}`),
//	}
//
// The native library must support the type and parameters. RawActionParam does
// not add typed result parsing or validate the parameters against the protocol.
// A nil value encodes as null. Converting a byte slice to RawActionParam shares
// its backing array; UnmarshalJSON copies its input.
type RawActionParam json.RawMessage

func (RawActionParam) isActionParam() {}

// MarshalJSON encodes the parameter as a JSON value, rejecting invalid JSON.
func (p RawActionParam) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.RawMessage(p))
}

// UnmarshalJSON validates and copies the parameter JSON without interpreting its fields.
func (p *RawActionParam) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*json.RawMessage)(p))
}

// RawRecognitionParam holds an uninterpreted JSON recognition parameter value.
// Recognition.UnmarshalJSON uses *RawRecognitionParam for unrecognized recognition types.
// It preserves parameter content, including large JSON integers, without converting
// it to Go values. JSON encoding may normalize whitespace and escape sequences.
//
// It can also be supplied wherever RecognitionParam is accepted. The native
// library must support the type and parameters; this does not add typed result
// parsing or validate parameters against the protocol. A nil value encodes as null.
// Converting a byte slice shares its backing array; UnmarshalJSON copies its input.
type RawRecognitionParam json.RawMessage

func (RawRecognitionParam) isRecognitionParam() {}

// MarshalJSON encodes the parameter as a JSON value, rejecting invalid JSON.
func (p RawRecognitionParam) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.RawMessage(p))
}

// UnmarshalJSON validates and copies the parameter JSON without interpreting its fields.
func (p *RawRecognitionParam) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*json.RawMessage)(p))
}
