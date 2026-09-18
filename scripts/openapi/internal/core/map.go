package core

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

type TrackState uint8

const (
	TrackUnused TrackState = iota
	TrackApplied
	TrackRedundant
)

type TrackedMap[K comparable, V any] struct {
	registry datautils.Map[K, V]
	states   datautils.Map[K, TrackState]
}

func NewTrackedMap[K comparable, V any](registry map[K]V) TrackedMap[K, V] {
	states := make(datautils.Map[K, TrackState])
	for key := range registry {
		states[key] = TrackUnused
	}

	return TrackedMap[K, V]{
		registry: registry,
		states:   states,
	}
}

func (m TrackedMap[K, V]) Get(key K) (V, bool) {
	value, ok := m.registry[key]
	if ok {
		m.states[key] = TrackApplied
	}

	return value, ok
}

func (m TrackedMap[K, V]) MarkRedundant(key K) {
	m.states[key] = TrackRedundant
}

func (m TrackedMap[K, V]) GetRedundantKeys() []K {
	keys := make([]K, 0)

	for key, state := range m.states {
		if state == TrackRedundant || state == TrackUnused {
			keys = append(keys, key)
		}
	}

	return keys
}

func (m TrackedMap[K, V]) GetAppliedKeys() []K {
	keys := make([]K, 0)

	for key, state := range m.states {
		if state == TrackApplied {
			keys = append(keys, key)
		}
	}

	return keys
}
