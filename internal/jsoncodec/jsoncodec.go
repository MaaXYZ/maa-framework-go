// Package jsoncodec provides the process-wide JSON codec used to serialize
// and deserialize the JSON that crosses the MaaFramework native boundary.
// The defaults are encoding/json's Marshal and Unmarshal; the maa package
// exposes the codec through SetJSONEncoder, SetJSONDecoder, GetJSONEncoder,
// GetJSONDecoder, and ResetJSONCodec. The encoder and decoder may be replaced
// concurrently with use; lookups are atomic.
package jsoncodec

import (
	"encoding/json"
	"sync/atomic"
)

// Encoder is the function type that serializes a value into JSON.
type Encoder func(v any) ([]byte, error)

// Decoder is the function type that deserializes JSON into a value.
type Decoder func(data []byte, v any) error

var (
	defaultEncoder Encoder = json.Marshal
	defaultDecoder Decoder = json.Unmarshal

	encoderStore atomic.Value
	decoderStore atomic.Value
)

func init() {
	encoderStore.Store(defaultEncoder)
	decoderStore.Store(defaultDecoder)
}

// SetEncoder atomically replaces the process-wide encoder.
// It panics if encoder is nil.
func SetEncoder(encoder Encoder) {
	if encoder == nil {
		panic("json encoder cannot be nil")
	}
	encoderStore.Store(encoder)
}

// SetDecoder atomically replaces the process-wide decoder.
// It panics if decoder is nil.
func SetDecoder(decoder Decoder) {
	if decoder == nil {
		panic("json decoder cannot be nil")
	}
	decoderStore.Store(decoder)
}

// GetEncoder returns the current process-wide encoder.
func GetEncoder() Encoder {
	return encoderStore.Load().(Encoder)
}

// GetDecoder returns the current process-wide decoder.
func GetDecoder() Decoder {
	return decoderStore.Load().(Decoder)
}

// Reset restores the default encoding/json Marshal and Unmarshal.
func Reset() {
	encoderStore.Store(defaultEncoder)
	decoderStore.Store(defaultDecoder)
}

// Marshal serializes v with the current process-wide encoder.
func Marshal(v any) ([]byte, error) {
	return GetEncoder()(v)
}

// Unmarshal deserializes data into v with the current process-wide decoder.
func Unmarshal(data []byte, v any) error {
	return GetDecoder()(data, v)
}
