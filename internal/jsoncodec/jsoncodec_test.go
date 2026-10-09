package jsoncodec

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetForTest(t *testing.T) {
	t.Helper()
	Reset()
	t.Cleanup(Reset)
}

func TestDefaultMatchesEncodingJSON(t *testing.T) {
	resetForTest(t)

	type sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	in := sample{Name: "alice", Age: 12}

	got, err := Marshal(in)
	require.NoError(t, err)

	want, err := json.Marshal(in)
	require.NoError(t, err)
	require.Equal(t, want, got)

	var decoded sample
	require.NoError(t, Unmarshal(got, &decoded))
	require.Equal(t, in, decoded)
}

func TestSetEncoderNilPanics(t *testing.T) {
	resetForTest(t)

	require.PanicsWithValue(t, "json encoder cannot be nil", func() {
		SetEncoder(nil)
	})
}

func TestSetDecoderNilPanics(t *testing.T) {
	resetForTest(t)

	require.PanicsWithValue(t, "json decoder cannot be nil", func() {
		SetDecoder(nil)
	})
}

func TestGetEncoderReturnsLastSet(t *testing.T) {
	resetForTest(t)

	enc := func(v any) ([]byte, error) {
		return []byte(`"encoder-stub"`), nil
	}
	SetEncoder(enc)

	got, err := GetEncoder()(1)
	require.NoError(t, err)
	require.Equal(t, []byte(`"encoder-stub"`), got)
}

func TestGetDecoderReturnsLastSet(t *testing.T) {
	resetForTest(t)

	dec := func(data []byte, v any) error {
		out, ok := v.(*string)
		if !ok {
			return errors.New("unexpected target")
		}
		*out = "decoder-stub"
		return nil
	}
	SetDecoder(dec)

	var out string
	require.NoError(t, GetDecoder()([]byte(`{}`), &out))
	require.Equal(t, "decoder-stub", out)
}

func TestResetRestoresEncodingJSONDefaults(t *testing.T) {
	resetForTest(t)

	SetEncoder(func(v any) ([]byte, error) {
		return []byte(`"not-json-marshal"`), nil
	})
	SetDecoder(func(data []byte, v any) error {
		return errors.New("decoder replaced")
	})

	Reset()

	type sample struct {
		Name string `json:"name"`
	}
	in := sample{Name: "alice"}

	got, err := Marshal(in)
	require.NoError(t, err)
	want, err := json.Marshal(in)
	require.NoError(t, err)
	require.Equal(t, want, got)

	var decoded sample
	require.NoError(t, Unmarshal(want, &decoded))
	require.Equal(t, in, decoded)
}

func TestMarshalUnmarshalDispatchThroughConfiguredCodec(t *testing.T) {
	resetForTest(t)

	encoderCalled := false
	SetEncoder(func(v any) ([]byte, error) {
		encoderCalled = true
		return []byte(`{"encoded":true}`), nil
	})

	decoderCalled := false
	SetDecoder(func(data []byte, v any) error {
		decoderCalled = true
		out, ok := v.(*string)
		if !ok {
			return errors.New("unexpected target")
		}
		*out = string(data)
		return nil
	})

	got, err := Marshal(1)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"encoded":true}`), got)
	require.True(t, encoderCalled)

	var out string
	require.NoError(t, Unmarshal([]byte("decoded"), &out))
	require.Equal(t, "decoded", out)
	require.True(t, decoderCalled)
}

func TestConcurrentSetGetAndUse(t *testing.T) {
	resetForTest(t)

	const workers = 8
	const rounds = 100

	var wg sync.WaitGroup
	wg.Add(workers)
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < rounds; j++ {
				SetEncoder(func(v any) ([]byte, error) {
					return json.Marshal(v)
				})
				SetDecoder(func(data []byte, v any) error {
					return json.Unmarshal(data, v)
				})

				if GetEncoder() == nil || GetDecoder() == nil {
					errCh <- errors.New("nil codec after Set")
					return
				}

				b, err := Marshal(map[string]int{"x": 1})
				if err != nil {
					errCh <- err
					return
				}

				var out map[string]int
				if err := Unmarshal(b, &out); err != nil {
					errCh <- err
					return
				}
				if out["x"] != 1 {
					errCh <- errors.New("unexpected decoded value")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}
