package store

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_ZeroValueGetForAbsentHandle(t *testing.T) {
	taskers := New[TaskerStoreValue]()
	taskers.Lock()
	value := taskers.Get(42)
	taskers.Unlock()
	require.Nil(t, value.SinkIDToEventCallbackID)
	require.Nil(t, value.ContextSinkIDToEventCallbackID)

	// The destroy path ranges over the maps of a freshly read value, so the
	// zero value must stay safe to iterate.
	for range value.SinkIDToEventCallbackID {
		t.Fatal("absent handle must yield no entries")
	}

	numbers := New[int]()
	numbers.Lock()
	number := numbers.Get(7)
	numbers.Unlock()
	require.Zero(t, number)
}

func TestStore_SetGetDelRoundTrip(t *testing.T) {
	stored := CtrlStoreValue{
		SinkIDToEventCallbackID:     map[int64]uint64{1: 10},
		CustomControllerCallbacksID: 5,
	}
	ctrls := New[CtrlStoreValue]()
	ctrls.Lock()
	ctrls.Set(7, stored)
	loaded := ctrls.Get(7)
	ctrls.Unlock()
	require.Equal(t, stored, loaded)

	ctrls.Lock()
	ctrls.Del(7)
	ctrls.Del(7) // deleting an absent handle stays a no-op
	after := ctrls.Get(7)
	ctrls.Unlock()
	require.Zero(t, after)
}

func TestStore_ConcurrentLockedAccess(t *testing.T) {
	const workers, increments = 8, 100
	counts := New[int]()
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range increments {
				counts.Lock()
				counts.Set(1, counts.Get(1)+1)
				counts.Unlock()
			}
		}()
	}
	wg.Wait()
	counts.Lock()
	defer counts.Unlock()
	require.Equal(t, workers*increments, counts.Get(1))
}

func TestStore_UpdateIsAtomic(t *testing.T) {
	const workers, updates = 8, 100
	counts := New[int]()
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range updates {
				counts.Update(1, func(value *int) { *value++ })
			}
		}()
	}
	wg.Wait()
	counts.Lock()
	defer counts.Unlock()
	require.Equal(t, workers*updates, counts.Get(1))
}

func TestStore_UpdateAbsentKeyStartsFromZero(t *testing.T) {
	resources := New[ResStoreValue]()
	resources.Update(9, func(value *ResStoreValue) {
		require.Nil(t, value.CustomRecognizersCallbackID)
		value.CustomRecognizersCallbackID = map[string]uint64{"reco": 1}
	})
	resources.Lock()
	stored := resources.Get(9)
	resources.Unlock()
	require.Equal(t, map[string]uint64{"reco": 1}, stored.CustomRecognizersCallbackID)
}

func TestStore_UpdateSavesModifiedCopy(t *testing.T) {
	taskers := New[TaskerStoreValue]()
	taskers.Update(3, func(value *TaskerStoreValue) {
		value.SinkIDToEventCallbackID = map[int64]uint64{5: 50}
	})
	taskers.Lock()
	stored := taskers.Get(3)
	taskers.Unlock()
	require.Equal(t, map[int64]uint64{5: 50}, stored.SinkIDToEventCallbackID)
}
