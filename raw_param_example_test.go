package maa_test

import (
	"encoding/json"
	"fmt"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

func ExampleRawActionParam() {
	action := &maa.Action{
		Type:  maa.ActionType("FutureAction"),
		Param: maa.RawActionParam(`{"value":9007199254740993}`),
	}
	encoded, err := json.Marshal(action)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))
	// Output: {"type":"FutureAction","param":{"value":9007199254740993}}
}

func ExampleRawRecognitionParam() {
	var recognition maa.Recognition
	err := json.Unmarshal([]byte(`{"type":"FutureMatcher","param":{"new_option":true}}`), &recognition)
	if err != nil {
		panic(err)
	}
	param := recognition.Param.(*maa.RawRecognitionParam)
	fmt.Println(recognition.Type)
	fmt.Println(string(*param))
	// Output:
	// FutureMatcher
	// {"new_option":true}
}
