package maa

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShellParam_UnmarshalJSON_Overflow(t *testing.T) {
	for _, input := range []string{
		`{"cmd":"new","shell_timeout":9223372036854775807}`,
		`{"cmd":"new","shell_timeout":-9223372036854775808}`,
	} {
		timeout := time.Second
		original := ShellParam{Cmd: "old", ShellTimeout: &timeout}
		got := original
		require.ErrorContains(t, unmarshalJSON([]byte(input), &got), "time.Duration range")
		require.Equal(t, original, got)
	}
}

func TestAction_AutoUpJSON(t *testing.T) {
	for _, actionType := range []ActionType{ActionTypeTouchDown, ActionTypeKeyDown} {
		for _, value := range []string{"", "true", "false"} {
			name := value
			if name == "" {
				name = "unspecified"
			}
			t.Run(string(actionType)+"/"+name, func(t *testing.T) {
				param := `{"key":65}`
				if actionType == ActionTypeTouchDown {
					param = `{"contact":1,"target_offset":[3,4,1,1]}`
				}
				if value != "" {
					param = param[:len(param)-1] + `,"auto_up":` + value + `}`
				}
				input := fmt.Sprintf(`{"type":%q,"param":%s}`, actionType, param)
				var action Action
				require.NoError(t, unmarshalJSON([]byte(input), &action))
				var autoUp *bool
				switch p := action.Param.(type) {
				case *TouchDownParam:
					autoUp = p.AutoUp
				case *KeyDownParam:
					autoUp = p.AutoUp
				default:
					t.Fatalf("unexpected parameter type %T", action.Param)
				}
				if value == "" {
					require.Nil(t, autoUp)
				} else {
					require.NotNil(t, autoUp)
					require.Equal(t, value == "true", *autoUp)
				}
				output, err := marshalJSON(action)
				require.NoError(t, err)
				require.JSONEq(t, input, string(output))
			})
		}
	}
}

func TestResource_OverridePipeline_AutoUp(t *testing.T) {
	for _, actionType := range []ActionType{ActionTypeTouchDown, ActionTypeKeyDown} {
		t.Run(string(actionType), func(t *testing.T) {
			res := createResource(t)
			t.Cleanup(func() { require.NoError(t, res.Destroy()) })
			makeAction := func(autoUp *bool) *Action {
				if actionType == ActionTypeTouchDown {
					return ActTouchDown(TouchDownParam{Contact: 1, AutoUp: autoUp})
				}
				return &Action{Type: ActionTypeKeyDown, Param: &KeyDownParam{Key: 65, AutoUp: autoUp}}
			}
			enabled, disabled := true, false
			for _, step := range []struct {
				name  string
				value *bool
				want  bool
			}{
				{"enable", &enabled, true},
				{"inherit enabled", nil, true},
				{"disable", &disabled, false},
				{"inherit disabled", nil, false},
			} {
				require.NoError(t, res.OverridePipeline(map[string]*Node{
					"AutoUp": {Action: makeAction(step.value)},
				}), step.name)
				node, err := res.GetNode("AutoUp")
				require.NoError(t, err, step.name)
				require.NotNil(t, node.Action)
				var autoUp *bool
				switch p := node.Action.Param.(type) {
				case *TouchDownParam:
					autoUp = p.AutoUp
				case *KeyDownParam:
					autoUp = p.AutoUp
				default:
					t.Fatalf("unexpected parameter type %T", node.Action.Param)
				}
				require.NotNil(t, autoUp, step.name)
				require.Equal(t, step.want, *autoUp, step.name)
			}
		})
	}
}
