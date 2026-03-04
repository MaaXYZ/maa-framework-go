package maa

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetJSONCodecForTest(t *testing.T) {
	t.Helper()
	ResetJSONCodec()
	t.Cleanup(ResetJSONCodec)
}

func TestJSONCodec_DefaultMatchesEncodingJSON(t *testing.T) {
	resetJSONCodecForTest(t)

	type sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	in := sample{Name: "alice", Age: 12}

	got, err := marshalJSON(in)
	require.NoError(t, err)

	want, err := json.Marshal(in)
	require.NoError(t, err)
	require.Equal(t, want, got)

	var decoded sample
	require.NoError(t, unmarshalJSON(got, &decoded))
	require.Equal(t, in, decoded)
}

func TestSetJSONEncoderAndDecoder(t *testing.T) {
	resetJSONCodecForTest(t)

	encoderCalled := false
	SetJSONEncoder(func(v any) ([]byte, error) {
		encoderCalled = true
		_, ok := v.(int)
		require.True(t, ok)
		return []byte(`"encoded"`), nil
	})

	decoderCalled := false
	SetJSONDecoder(func(data []byte, v any) error {
		decoderCalled = true
		out, ok := v.(*string)
		require.True(t, ok)
		*out = string(data)
		return nil
	})

	got, err := marshalJSON(1)
	require.NoError(t, err)
	require.Equal(t, []byte(`"encoded"`), got)
	require.True(t, encoderCalled)

	var out string
	require.NoError(t, unmarshalJSON([]byte("decoded"), &out))
	require.Equal(t, "decoded", out)
	require.True(t, decoderCalled)
}

func TestSetJSONEncoderNilPanics(t *testing.T) {
	resetJSONCodecForTest(t)
	require.Panics(t, func() {
		SetJSONEncoder(nil)
	})
}

func TestSetJSONDecoderNilPanics(t *testing.T) {
	resetJSONCodecForTest(t)
	require.Panics(t, func() {
		SetJSONDecoder(nil)
	})
}

func TestWithJSONEncoderNilPanics(t *testing.T) {
	require.Panics(t, func() {
		_ = WithJSONEncoder(nil)
	})
}

func TestWithJSONDecoderNilPanics(t *testing.T) {
	require.Panics(t, func() {
		_ = WithJSONDecoder(nil)
	})
}

func TestJSONCodecLastWriteWins(t *testing.T) {
	resetJSONCodecForTest(t)

	SetJSONEncoder(func(v any) ([]byte, error) {
		return []byte(`"old"`), nil
	})
	SetJSONEncoder(func(v any) ([]byte, error) {
		return []byte(`"new"`), nil
	})

	got, err := marshalJSON("x")
	require.NoError(t, err)
	require.Equal(t, []byte(`"new"`), got)
}

func TestResetJSONCodecRestoresDefault(t *testing.T) {
	resetJSONCodecForTest(t)

	SetJSONEncoder(func(v any) ([]byte, error) {
		return nil, errors.New("boom")
	})

	_, err := marshalJSON("x")
	require.Error(t, err)

	ResetJSONCodec()

	type sample struct {
		Value string `json:"value"`
	}
	got, err := marshalJSON(sample{Value: "ok"})
	require.NoError(t, err)
	require.Equal(t, []byte(`{"value":"ok"}`), got)
}

func TestWithJSONEncoderAndDecoderOptionApply(t *testing.T) {
	enc := JSONEncoder(func(v any) ([]byte, error) {
		return []byte("enc"), nil
	})
	dec := JSONDecoder(func(data []byte, v any) error {
		return nil
	})

	cfg := initConfig{}
	WithJSONEncoder(enc)(&cfg)
	WithJSONDecoder(dec)(&cfg)

	require.NotNil(t, cfg.JSONEncoder)
	require.NotNil(t, cfg.JSONDecoder)

	gotEnc := *cfg.JSONEncoder
	gotDec := *cfg.JSONDecoder

	b, err := gotEnc(1)
	require.NoError(t, err)
	require.Equal(t, []byte("enc"), b)
	require.NoError(t, gotDec([]byte("x"), new(any)))
}

func TestGetJSONEncoderAndDecoder(t *testing.T) {
	resetJSONCodecForTest(t)

	enc := JSONEncoder(func(v any) ([]byte, error) {
		return []byte(`"ok"`), nil
	})
	dec := JSONDecoder(func(data []byte, v any) error {
		out := v.(*string)
		*out = "ok"
		return nil
	})

	SetJSONEncoder(enc)
	SetJSONDecoder(dec)

	gotEnc := GetJSONEncoder()
	gotDec := GetJSONDecoder()
	require.NotNil(t, gotEnc)
	require.NotNil(t, gotDec)

	b, err := gotEnc(1)
	require.NoError(t, err)
	require.Equal(t, []byte(`"ok"`), b)

	var out string
	require.NoError(t, gotDec([]byte(`"ignored"`), &out))
	require.Equal(t, "ok", out)
}

func TestJSONCodecConcurrentSetAndUse(t *testing.T) {
	resetJSONCodecForTest(t)

	const workers = 8
	const rounds = 100

	var wg sync.WaitGroup
	wg.Add(workers)
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < rounds; j++ {
				SetJSONEncoder(func(v any) ([]byte, error) {
					return json.Marshal(v)
				})
				SetJSONDecoder(func(data []byte, v any) error {
					return json.Unmarshal(data, v)
				})

				b, err := marshalJSON(map[string]int{"x": 1})
				if err != nil {
					errCh <- err
					return
				}

				var out map[string]int
				if err := unmarshalJSON(b, &out); err != nil {
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
