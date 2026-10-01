package maa

import "github.com/MaaXYZ/maa-framework-go/v4/internal/jsoncodec"

// JSONEncoder defines how values are serialized into JSON.
//
// Custom encoders must honor the encoding/json semantics used by this package,
// including json.Marshaler, "omitempty", and the "omitzero" option added in Go 1.24.
// Support for "omitzero" includes calling IsZero() methods. For this package's
// slice fields tagged "omitzero", nil must be omitted while non-nil empty slices
// must be encoded as [].
//
// These rules preserve pipeline defaults and inheritance. Emitting null instead
// of omitting a zero field may cause MaaFramework to reject the configuration.
type JSONEncoder func(v any) ([]byte, error)

// JSONDecoder defines how JSON is deserialized into values.
type JSONDecoder func(data []byte, v any) error

// SetJSONEncoder sets the global JSON encoder used by this package.
//
// The encoder must satisfy the compatibility requirements documented by
// [JSONEncoder], including support for "omitzero" and IsZero() methods.
func SetJSONEncoder(encoder JSONEncoder) {
	jsoncodec.SetEncoder(jsoncodec.Encoder(encoder))
}

// SetJSONDecoder sets the global JSON decoder used by this package.
func SetJSONDecoder(decoder JSONDecoder) {
	jsoncodec.SetDecoder(jsoncodec.Decoder(decoder))
}

// GetJSONEncoder returns the currently configured global JSON encoder.
func GetJSONEncoder() JSONEncoder {
	return JSONEncoder(jsoncodec.GetEncoder())
}

// GetJSONDecoder returns the currently configured global JSON decoder.
func GetJSONDecoder() JSONDecoder {
	return JSONDecoder(jsoncodec.GetDecoder())
}

// ResetJSONCodec resets the global JSON encoder and decoder to encoding/json defaults.
func ResetJSONCodec() {
	jsoncodec.Reset()
}

func marshalJSON(v any) ([]byte, error) {
	return jsoncodec.Marshal(v)
}

func unmarshalJSON(data []byte, v any) error {
	return jsoncodec.Unmarshal(data, v)
}
